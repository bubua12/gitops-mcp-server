# 三、功能模块设计

## 3.0 功能模块总览

```
┌─────────────────────────────────────────────────────────────────┐
│                     GitOps MCP Server                           │
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ M1 仓库   │ │ M2 Issue │ │ M3 PR &  │ │ M4 Code  │           │
│  │ 管理     │ │ 管理     │ │ Review   │ │ Intelligence│         │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
│                                                                 │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐           │
│  │ M5 Release│ │ M6 CI/CD │ │ M7 事件  │ │ M8 通知  │           │
│  │ 管理     │ │ 管理     │ │ 监控     │ │ 系统     │           │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘           │
│                                                                 │
│  ┌──────────────────────────────────────────────────────┐       │
│  │              M0 基础层：认证、配置、缓存、日志           │       │
│  └──────────────────────────────────────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

---

## M0：基础服务层

> 所有模块的公共底座，提供认证、配置、缓存、日志、健康检查等基础能力。

### 0.1 认证管理

| 能力 | 说明 |
|------|------|
| Token 管理 | 支持 GitHub PAT（Personal Access Token）、GitHub App、OAuth |
| 多账户支持 | 可配置多个 GitHub 账户/token，按仓库自动匹配 |
| Token 轮换 | 支持 Token 过期自动刷新（GitHub App 场景） |
| 权限检查 | 启动时校验 Token 权限范围，提醒缺失权限 |

### 0.2 配置管理

```yaml
# gitops-mcp.config.yaml 示例
server:
  transport: stdio          # stdio | sse | streamable-http
  port: 3000                # 仅 SSE/HTTP 模式生效

github:
  default_owner: "wangxl"   # 默认仓库所有者
  tokens:
    - name: "personal"
      token: "${GITHUB_TOKEN}"
      default: true
    - name: "org-bot"
      token: "${ORG_BOT_TOKEN}"
      orgs: ["empoworx"]

  # 关注的仓库列表
  watched_repos:
    - owner: "kubernetes"
      repo: "kubernetes"
      watch_issues: true
      watch_releases: true
      watch_security: true

  # Issue 模板
  issue_templates:
    bug:
      labels: ["bug", "triage"]
      assignees: ["wangxl"]
    feature:
      labels: ["enhancement"]

  # 通知配置
  notifications:
    channels:
      - type: "terminal"
        enabled: true
      - type: "webhook"
        url: "https://hooks.dingtalk.com/xxx"
        enabled: false
      - type: "email"
        smtp_host: "smtp.example.com"
        enabled: false

cache:
  backend: "memory"         # memory | redis | file
  ttl:
    repo_structure: "1h"
    file_content: "5m"
    issue_list: "2m"

logging:
  level: "info"             # debug | info | warn | error
  format: "json"            # json | text
  output: "stderr"          # stderr | file
  file_path: "./logs/gitops-mcp.log"
```

### 0.3 速率限制与缓存

| 策略 | 说明 |
|------|------|
| API 速率感知 | 自动跟踪 GitHub API 速率限制（5000/h），接近限制时降速 |
| 响应缓存 | 对高频读操作（仓库结构、文件内容）进行 TTL 缓存 |
| 批量合并 | 短时间内多个同类请求自动合并为批量 API 调用 |

---

## M1：仓库管理模块

> 提供仓库信息查询、结构浏览、文件读取等基础能力。对标 ZRead 的核心功能，并扩展为读写双向。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `search_repositories` | 搜索仓库 | `query`, `sort`, `order`, `per_page` | 仓库列表 |
| `get_repository` | 获取仓库详情 | `owner`, `repo` | 仓库元信息 |
| `list_repositories` | 列出用户的仓库 | `type`(all/owner/member), `sort`, `per_page` | 仓库列表 |
| `get_repo_structure` | 获取仓库目录树 | `owner`, `repo`, `ref`(分支/tag), `path`(子目录) | 树形文件列表 |
| `read_file` | 读取文件内容 | `owner`, `repo`, `path`, `ref` | 文件内容 + 元信息 |
| `search_code` | 搜索代码 | `query`, `owner`, `repo`, `language`, `path` | 匹配结果 |
| `get_file_history` | 获取文件修改历史 | `owner`, `repo`, `path`, `per_page` | commit 列表 |
| `compare_refs` | 对比两个 ref 的差异 | `owner`, `repo`, `base`, `head` | diff 结果 |

### 关键特性

- **智能缓存**：仓库结构缓存 1 小时，文件内容缓存 5 分钟，避免重复 API 调用
- **大文件处理**：超过 1MB 的文件自动截断并提示，支持指定行范围读取
- **多 ref 支持**：支持 branch、tag、commit SHA 作为 ref 参数
- **目录递归深度控制**：`depth` 参数控制目录树展开层级，默认 3 层

---

## M2：Issue 管理模块

> Issue 的全生命周期管理，是本产品的核心差异化模块之一。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_issues` | 列出 Issues | `owner`, `repo`, `state`, `labels`, `assignee`, `since`, `sort`, `per_page` | Issue 列表 |
| `get_issue` | 获取 Issue 详情 | `owner`, `repo`, `number` | Issue 详情（含评论） |
| `create_issue` | 创建 Issue | `owner`, `repo`, `title`, `body`, `labels`, `assignees`, `milestone` | 创建的 Issue |
| `update_issue` | 更新 Issue | `owner`, `repo`, `number`, `title?`, `body?`, `state?`, `labels?`, `assignees?` | 更新后的 Issue |
| `close_issue` | 关闭 Issue | `owner`, `repo`, `number`, `reason`(completed/not_planned/duplicate) | 操作结果 |
| `add_comment` | 添加评论 | `owner`, `repo`, `number`, `body` | 评论 |
| `list_comments` | 列出评论 | `owner`, `repo`, `number`, `since` | 评论列表 |
| `add_labels` | 添加标签 | `owner`, `repo`, `number`, `labels[]` | 操作结果 |
| `remove_labels` | 移除标签 | `owner`, `repo`, `number`, `labels[]` | 操作结果 |
| `search_issues` | 搜索 Issue | `query`(GitHub search syntax), `owner?`, `repo?` | 搜索结果 |
| `create_issue_from_template` | 从模板创建 | `owner`, `repo`, `template`(bug/feature/custom), `params` | 创建的 Issue |
| `batch_create_issues` | 批量创建 | `owner`, `repo`, `issues[]` | 创建结果列表 |

### 关键特性

- **模板系统**：支持预设 bug、feature 等模板，自动填充标签、指派
- **智能标签**：根据 Issue 内容自动推荐标签（基于关键词匹配规则）
- **批量操作**：支持一次创建/更新多个 Issue，适合迁移场景
- **关联追踪**：`close_issue` 时自动在评论中记录关闭原因和关联信息

---

## M3：Pull Request & Code Review 模块

> PR 全生命周期管理和 AI 辅助代码审查。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_pull_requests` | 列出 PR | `owner`, `repo`, `state`, `base`, `head`, `sort`, `per_page` | PR 列表 |
| `get_pull_request` | 获取 PR 详情 | `owner`, `repo`, `number` | PR 详情 |
| `get_pr_diff` | 获取 PR diff | `owner`, `repo`, `number` | diff 内容 |
| `get_pr_files` | 获取 PR 变更文件列表 | `owner`, `repo`, `number` | 文件列表 + 变更统计 |
| `get_pr_reviews` | 获取审查历史 | `owner`, `repo`, `number` | 审查列表 |
| `create_review` | 提交代码审查 | `owner`, `repo`, `number`, `event`(APPROVE/REQUEST_CHANGES/COMMENT), `body`, `comments[]` | 审查结果 |
| `create_pull_request` | 创建 PR | `owner`, `repo`, `title`, `body`, `head`, `base`, `draft?` | 创建的 PR |
| `merge_pull_request` | 合并 PR | `owner`, `repo`, `number`, `merge_method`(merge/squash/rebase) | 合并结果 |
| `close_pull_request` | 关闭 PR | `owner`, `repo`, `number` | 操作结果 |
| `request_reviewers` | 请求审查者 | `owner`, `repo`, `number`, `reviewers[]`, `team_reviewers[]` | 操作结果 |
| `auto_review` | AI 自动审查（高级） | `owner`, `repo`, `number`, `rules[]`, `severity_threshold` | 审查报告 |

### 关键特性

- **inline 评论**：`create_review` 支持对具体行提交 inline comment
- **自动审查规则**：`auto_review` 可配置规则集（安全扫描、代码规范、复杂度检查等）
- **审查摘要**：自动生成 PR 变更的结构化摘要（改了什么、影响范围、风险点）

---

## M4：Code Intelligence 模块

> 超越 ZRead 的"只读浏览"，提供代码理解和智能分析能力。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `search_docs` | 搜索仓库文档 | `owner`, `repo`, `query` | 文档片段列表 |
| `get_readme` | 获取 README | `owner`, `repo`, `ref?` | README 内容 |
| `get_contributing_guide` | 获取贡献指南 | `owner`, `repo` | CONTRIBUTING.md 内容 |
| `get_recent_commits` | 获取最近提交 | `owner`, `repo`, `since?`, `until?`, `path?`, `author?`, `per_page` | commit 列表 |
| `get_commit_detail` | 获取 commit 详情 | `owner`, `repo`, `sha` | commit 详情 + diff |
| `analyze_changes` | 分析变更影响 | `owner`, `repo`, `base`, `head` | 影响分析报告 |
| `suggest_issues` | 基于代码问题建议 Issue | `owner`, `repo`, `path?`, `focus`(security/perf/style) | 建议列表 |
| `get_dependency_graph` | 依赖分析 | `owner`, `repo` | 依赖关系 |

### 关键特性

- **变更影响分析**：`analyze_changes` 分析一次 commit/PR 影响了哪些模块，是否有破坏性变更
- **智能 Issue 建议**：`suggest_issues` 扫描代码后，基于规则自动生成 Issue 建议（不自动创建，需用户确认）
- **文档优先搜索**：`search_docs` 优先搜索 README、docs/、wiki 等文档区域

---

## M5：Release 管理模块

> 自动化 Release 流程，从 Tag 创建到 Release Notes 生成。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_tags` | 列出 Tags | `owner`, `repo`, `per_page` | tag 列表 |
| `create_tag` | 创建 Tag | `owner`, `repo`, `tag_name`, `target_commitish`, `message?` | 创建的 tag |
| `delete_tag` | 删除 Tag | `owner`, `repo`, `tag_name` | 操作结果 |
| `list_releases` | 列出 Releases | `owner`, `repo`, `per_page` | release 列表 |
| `get_release` | 获取 Release 详情 | `owner`, `repo`, `tag` | release 详情 |
| `create_release` | 创建 Release | `owner`, `repo`, `tag_name`, `name`, `body`, `draft?`, `prerelease?`, `generate_notes?` | 创建的 release |
| `update_release` | 更新 Release | `owner`, `repo`, `id`, `name?`, `body?`, `draft?` | 更新后的 release |
| `generate_release_notes` | 自动生成 Release Notes | `owner`, `repo`, `tag_name`, `previous_tag?`, `format`(markdown/json) | 生成的 notes |
| `publish_release` | 发布草稿 Release | `owner`, `repo`, `id` | 操作结果 |

### 关键特性

- **智能 Release Notes**：基于 conventional commits（feat/fix/docs/breaking）自动分类生成
- **自动生成变更摘要**：不只是 commit 列表，而是 AI 理解后的结构化变更说明
- **前一个 Tag 自动推断**：`generate_release_notes` 自动查找上一个 release tag 作为对比基准

---

## M6：CI/CD 管理模块

> GitHub Actions 工作流的查询、触发和监控。

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_workflows` | 列出工作流 | `owner`, `repo` | 工作流列表 |
| `get_workflow` | 获取工作流详情 | `owner`, `repo`, `workflow_id` | 工作流详情 |
| `list_workflow_runs` | 列出工作流运行记录 | `owner`, `repo`, `workflow_id?`, `branch?`, `status?`, `per_page` | 运行记录 |
| `get_workflow_run` | 获取运行详情 | `owner`, `repo`, `run_id` | 运行详情 |
| `trigger_workflow` | 触发工作流 | `owner`, `repo`, `workflow_id`, `ref`, `inputs?` | 触发结果 |
| `rerun_workflow` | 重新运行工作流 | `owner`, `repo`, `run_id` | 操作结果 |
| `cancel_workflow` | 取消工作流 | `owner`, `repo`, `run_id` | 操作结果 |
| `get_job_logs` | 获取 Job 日志 | `owner`, `repo`, `job_id`, `tail?(行数)` | 日志内容 |
| `list_artifacts` | 列出构建产物 | `owner`, `repo`, `run_id` | 产物列表 |
| `download_artifact` | 下载构建产物 | `owner`, `repo`, `artifact_id` | 产物内容 |

### 关键特性

- **失败自动重试**：可配置失败后自动重试次数和间隔
- **日志智能分析**：`get_job_logs` 支持只获取最后 N 行，并自动提取错误信息
- **状态汇总**：一次调用获取所有工作流的最近状态，适合巡检场景

---

## M7：事件监控模块（Webhook + Polling）

> 实时或准实时监控 GitHub 事件，是与纯 CLI 工具的核心差异。

### 7.1 两种监控模式

```
模式 A：Webhook 模式（实时）
┌──────────────┐    POST     ┌──────────────────┐    推送    ┌──────────┐
│   GitHub     │ ─────────→ │  GitOps MCP      │ ────────→ │  通知渠道 │
│   Webhook    │  events    │  Webhook Server  │           │          │
└──────────────┘            └──────────────────┘           └──────────┘
延迟：< 1秒
前提：需要公网可访问的回调 URL

模式 B：Polling 模式（准实时）
┌──────────────┐   定时请求   ┌──────────────────┐   对比    ┌──────────┐
│   GitHub     │ ←───────── │  GitOps MCP      │ ────────→ │  通知渠道 │
│   API        │  polling    │  Poller          │  diff     │          │
└──────────────┘            └──────────────────┘           └──────────┘
延迟：可配置（默认 5 分钟）
前提：无需公网 URL，适合本地开发
```

### 7.2 监控规则配置

```yaml
# 监控规则示例
monitors:
  - name: "my-issues"
    type: "polling"            # polling | webhook
    interval: "5m"             # polling 间隔
    repos:
      - owner: "kubernetes"
        repo: "kubernetes"
      - owner: "prometheus"
        repo: "prometheus"
    filters:
      - type: "issue_comment"
        condition: "author != me"   # 只关注别人的回复
      - type: "issue_closed"
        condition: "closer != me"   # 别人关闭了我的 issue
      - type: "new_release"         # 关注新版本
      - type: "security_advisory"   # 安全公告
    notify:
      channels: ["terminal"]

  - name: "repo-health"
    type: "polling"
    interval: "1h"
    repos:
      - owner: "wangxl"
        repo: "*"
    filters:
      - type: "ci_failure"
      - type: "new_issue"
      - type: "pr_review_requested"
    notify:
      channels: ["terminal", "webhook"]
```

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_monitors` | 列出监控规则 | — | 规则列表 |
| `add_monitor` | 添加监控规则 | `name`, `type`, `repos[]`, `filters[]`, `interval?`, `notify` | 创建的规则 |
| `remove_monitor` | 移除监控规则 | `name` | 操作结果 |
| `pause_monitor` | 暂停监控 | `name` | 操作结果 |
| `resume_monitor` | 恢复监控 | `name` | 操作结果 |
| `get_notifications` | 获取未读通知 | `since?`, `read?`(true/false) | 通知列表 |
| `mark_notification_read` | 标记已读 | `id` | 操作结果 |
| `get_event_log` | 获取事件日志 | `monitor?`, `since?`, `per_page` | 事件列表 |

### 关键特性

- **双模式并存**：本地开发用 Polling，生产环境用 Webhook
- **灵活过滤**：支持事件类型 + 条件表达式的组合过滤
- **去重防抖**：同一事件短时间内不重复通知
- **事件回溯**：所有事件都有日志，支持事后查看

---

## M8：通知系统

> 统一的通知分发层，将各模块产生的通知推送到不同渠道。

### 支持的通知渠道

| 渠道 | 状态 | 说明 |
|------|------|------|
| **Terminal** | ✅ 优先实现 | 通过 MCP session 直接输出到终端 |
| **Webhook** | ✅ 优先实现 | 通用 Webhook，支持钉钉、企微、飞书、Slack 等 |
| **Email** | 📋 后续实现 | SMTP 邮件通知 |
| **GitHub Discussions** | 📋 后续实现 | 在指定 Discussion 分类下创建通知 |
| **自定义脚本** | 📋 后续实现 | 调用自定义脚本/命令 |

### Tools 列表

| Tool | 描述 | 参数 | 返回值 |
|------|------|------|--------|
| `list_notification_channels` | 列出通知渠道 | — | 渠道列表 |
| `add_notification_channel` | 添加渠道 | `name`, `type`, `config` | 创建的渠道 |
| `send_notification` | 手动发送通知 | `channel`, `title`, `body`, `level`(info/warn/error) | 发送结果 |
| `test_notification` | 测试渠道 | `channel` | 测试结果 |

### 通知格式模板

```json
{
  "title": "[GitOps] 新评论: kubernetes/kubernetes#12345",
  "body": "@someone 回复了你提出的 Issue:\n\n> 你的原始描述...\n\n> 他们的回复...\n\n链接: https://github.com/...",
  "level": "info",
  "timestamp": "2026-06-25T10:30:00Z",
  "source": {
    "type": "issue_comment",
    "repo": "kubernetes/kubernetes",
    "number": 12345
  }
}
```
