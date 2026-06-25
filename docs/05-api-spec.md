# 五、MCP Tool 接口规范

## 5.1 MCP Tool Schema 规范

所有 Tool 遵循 JSON Schema 标准定义，兼容 MCP 协议规范。

### 通用约定

| 规则 | 说明 |
|------|------|
| 参数命名 | 使用 `snake_case`，与 GitHub REST API 保持一致 |
| owner/repo | 大部分 Tool 必填；`default_owner` 配置后可省略 owner |
| 分页 | 统一使用 `per_page`（默认 30，最大 100）和 `page` 参数 |
| 返回格式 | 统一返回 Markdown 文本（人类可读），部分返回结构化 JSON |
| 错误处理 | 统一错误格式 `{ "error": "描述", "code": "GITHUB_API_ERROR" }` |

## 5.2 完整 Tool 清单

### M1：仓库管理（8 个 Tool）

```yaml
# 1. search_repositories
name: search_repositories
description: "搜索 GitHub 仓库。支持按关键词、语言、star 数等条件搜索。"
inputSchema:
  type: object
  properties:
    query:
      type: string
      description: "搜索关键词，支持 GitHub 搜索语法，如 'language:go stars:>1000'"
    sort:
      type: string
      enum: [stars, forks, help-wanted-issues, updated]
      description: "排序方式"
    order:
      type: string
      enum: [asc, desc]
      default: desc
    per_page:
      type: integer
      default: 30
      maximum: 100
  required: [query]

# 2. get_repository
name: get_repository
description: "获取指定仓库的详细信息，包括描述、语言、star/fork 数、默认分支等。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
      description: "仓库所有者（如未指定则使用配置中的 default_owner）"
    repo:
      type: string
      description: "仓库名称"
  required: [repo]

# 3. list_repositories
name: list_repositories
description: "列出当前认证用户的仓库列表。"
inputSchema:
  type: object
  properties:
    type:
      type: string
      enum: [all, owner, member, public, private]
      default: all
    sort:
      type: string
      enum: [created, updated, pushed, full_name]
      default: full_name
    per_page:
      type: integer
      default: 30
      maximum: 100

# 4. get_repo_structure
name: get_repo_structure
description: "获取仓库的目录结构树。支持指定分支/tag 和子目录路径。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    ref:
      type: string
      description: "分支名、tag 或 commit SHA（默认：默认分支）"
    path:
      type: string
      description: "子目录路径（默认：根目录）"
    depth:
      type: integer
      default: 3
      description: "递归深度"
  required: [repo]

# 5. read_file
name: read_file
description: "读取仓库中的文件内容。支持指定行范围。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    path:
      type: string
      description: "文件路径"
    ref:
      type: string
      description: "分支/tag/SHA"
    line_start:
      type: integer
      description: "起始行号（从 1 开始）"
    line_end:
      type: integer
      description: "结束行号"
  required: [repo, path]

# 6. search_code
name: search_code
description: "在仓库中搜索代码内容。支持正则表达式。"
inputSchema:
  type: object
  properties:
    query:
      type: string
      description: "搜索关键词或正则表达式"
    owner:
      type: string
    repo:
      type: string
    language:
      type: string
      description: "限定语言"
    path:
      type: string
      description: "限定路径前缀"
    per_page:
      type: integer
      default: 30
  required: [query]

# 7. get_file_history
name: get_file_history
description: "获取指定文件的修改历史（commit 列表）。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    path:
      type: string
    per_page:
      type: integer
      default: 30
  required: [repo, path]

# 8. compare_refs
name: compare_refs
description: "对比两个 ref（分支/tag/SHA）之间的差异。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    base:
      type: string
      description: "基准 ref"
    head:
      type: string
      description: "对比 ref"
  required: [repo, base, head]
```

### M2：Issue 管理（10 个 Tool）

```yaml
# 1. list_issues
name: list_issues
description: "列出仓库的 Issues。支持按状态、标签、指派人过滤。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    state:
      type: string
      enum: [open, closed, all]
      default: open
    labels:
      type: string
      description: "逗号分隔的标签名"
    assignee:
      type: string
    since:
      type: string
      description: "ISO 8601 时间，只返回该时间之后更新的 Issue"
    sort:
      type: string
      enum: [created, updated, comments]
      default: created
    direction:
      type: string
      enum: [asc, desc]
      default: desc
    per_page:
      type: integer
      default: 30
  required: [repo]

# 2. get_issue
name: get_issue
description: "获取 Issue 详情，包含全部评论。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
      description: "Issue 编号"
  required: [repo, number]

# 3. create_issue
name: create_issue
description: "创建新 Issue。支持设置标签、指派人、关联里程碑。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    title:
      type: string
    body:
      type: string
      description: "Issue 描述，支持 Markdown"
    labels:
      type: array
      items:
        type: string
    assignees:
      type: array
      items:
        type: string
    milestone:
      type: integer
      description: "里程碑编号"
  required: [repo, title]

# 4. update_issue
name: update_issue
description: "更新 Issue 的标题、描述、状态、标签等。所有字段均可选。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    title:
      type: string
    body:
      type: string
    state:
      type: string
      enum: [open, closed]
    state_reason:
      type: string
      enum: [completed, not_planned, duplicate]
    labels:
      type: array
      items:
        type: string
    assignees:
      type: array
      items:
        type: string
    milestone:
      type: integer
  required: [repo, number]

# 5. close_issue
name: close_issue
description: "关闭 Issue。支持指定关闭原因。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    reason:
      type: string
      enum: [completed, not_planned, duplicate]
      default: completed
    comment:
      type: string
      description: "关闭时自动添加的评论（可选）"
  required: [repo, number]

# 6. add_comment
name: add_comment
description: "对 Issue 添加评论。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    body:
      type: string
      description: "评论内容，支持 Markdown"
  required: [repo, number, body]

# 7. list_comments
name: list_comments
description: "列出 Issue 的评论。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    since:
      type: string
      description: "只返回该时间之后的评论"
    per_page:
      type: integer
      default: 30
  required: [repo, number]

# 8. add_labels
name: add_labels
description: "为 Issue 添加标签。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    labels:
      type: array
      items:
        type: string
  required: [repo, number, labels]

# 9. remove_labels
name: remove_labels
description: "移除 Issue 的标签。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    labels:
      type: array
      items:
        type: string
  required: [repo, number, labels]

# 10. search_issues
name: search_issues
description: "使用 GitHub 搜索语法搜索 Issue。"
inputSchema:
  type: object
  properties:
    query:
      type: string
      description: "搜索查询，支持 GitHub Issue 搜索语法"
    owner:
      type: string
    repo:
      type: string
    sort:
      type: string
      enum: [created, updated, comments]
    per_page:
      type: integer
      default: 30
  required: [query]
```

### M3：PR & Review（7 个 Tool）

```yaml
# 1. list_pull_requests
name: list_pull_requests
description: "列出仓库的 Pull Requests。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    state:
      type: string
      enum: [open, closed, all]
      default: open
    base:
      type: string
      description: "目标分支过滤"
    head:
      type: string
      description: "源分支过滤"
    sort:
      type: string
      enum: [created, updated, popularity, long-running]
      default: created
    per_page:
      type: integer
      default: 30
  required: [repo]

# 2. get_pull_request
name: get_pull_request
description: "获取 PR 详情，包括审查状态、CI 状态、合并状态。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
  required: [repo, number]

# 3. get_pr_diff
name: get_pr_diff
description: "获取 PR 的代码 diff。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
  required: [repo, number]

# 4. get_pr_files
name: get_pr_files
description: "获取 PR 中变更的文件列表及每个文件的变更统计。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    per_page:
      type: integer
      default: 30
  required: [repo, number]

# 5. create_pull_request
name: create_pull_request
description: "创建 Pull Request。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    title:
      type: string
    body:
      type: string
    head:
      type: string
      description: "源分支"
    base:
      type: string
      description: "目标分支"
    draft:
      type: boolean
      default: false
  required: [repo, title, head, base]

# 6. create_review
name: create_review
description: "对 PR 提交代码审查。支持 APPROVE / REQUEST_CHANGES / COMMENT。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    event:
      type: string
      enum: [APPROVE, REQUEST_CHANGES, COMMENT]
    body:
      type: string
      description: "审查总评"
    comments:
      type: array
      description: "inline 评论列表"
      items:
        type: object
        properties:
          path:
            type: string
            description: "文件路径"
          line:
            type: integer
            description: "行号"
          body:
            type: string
            description: "评论内容"
  required: [repo, number, event]

# 7. merge_pull_request
name: merge_pull_request
description: "合并 Pull Request。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    number:
      type: integer
    merge_method:
      type: string
      enum: [merge, squash, rebase]
      default: merge
    commit_title:
      type: string
    commit_message:
      type: string
  required: [repo, number]
```

### M5：Release 管理（5 个核心 Tool）

```yaml
# 1. list_tags
name: list_tags
description: "列出仓库的 Tags。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    per_page:
      type: integer
      default: 30
  required: [repo]

# 2. create_tag
name: create_tag
description: "创建轻量 Tag 或附注 Tag（annotated tag）。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    tag_name:
      type: string
    target_commitish:
      type: string
      description: "目标 commit SHA 或分支名"
    message:
      type: string
      description: "Tag 消息（附注 Tag）"
  required: [repo, tag_name]

# 3. list_releases
name: list_releases
description: "列出仓库的 Releases。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    per_page:
      type: integer
      default: 30
  required: [repo]

# 4. create_release
name: create_release
description: "创建 Release。可自动生成 Release Notes。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    tag_name:
      type: string
    name:
      type: string
      description: "Release 标题"
    body:
      type: string
      description: "Release 描述"
    draft:
      type: boolean
      default: false
    prerelease:
      type: boolean
      default: false
    generate_notes:
      type: boolean
      default: false
      description: "是否让 GitHub 自动生成 Release Notes"
  required: [repo, tag_name]

# 5. generate_release_notes
name: generate_release_notes
description: "基于 conventional commits 自动分类生成 Release Notes。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    tag_name:
      type: string
    previous_tag:
      type: string
      description: "对比基准 tag（默认自动推断上一个 release）"
    format:
      type: string
      enum: [markdown, json]
      default: markdown
  required: [repo, tag_name]
```

### M6：CI/CD 管理（6 个核心 Tool）

```yaml
# 1. list_workflows
name: list_workflows
description: "列出仓库的 GitHub Actions 工作流。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
  required: [repo]

# 2. list_workflow_runs
name: list_workflow_runs
description: "列出工作流的运行记录。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    workflow_id:
      type: string
      description: "工作流 ID 或文件名"
    branch:
      type: string
    status:
      type: string
      enum: [queued, in_progress, completed, success, failure, cancelled]
    per_page:
      type: integer
      default: 30
  required: [repo]

# 3. trigger_workflow
name: trigger_workflow
description: "手动触发工作流运行。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    workflow_id:
      type: string
    ref:
      type: string
      description: "触发的分支或 tag"
    inputs:
      type: object
      description: "workflow_dispatch 输入参数"
  required: [repo, workflow_id, ref]

# 4. rerun_workflow
name: rerun_workflow
description: "重新运行失败的工作流。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    run_id:
      type: integer
  required: [repo, run_id]

# 5. get_job_logs
name: get_job_logs
description: "获取 CI Job 的日志。可只获取最后 N 行。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
    job_id:
      type: integer
    tail:
      type: integer
      description: "只返回最后 N 行（默认 100）"
      default: 100
  required: [repo, job_id]

# 6. get_workflow_summary
name: get_workflow_summary
description: "一次调用获取所有工作流的最近运行状态汇总。适合巡检。"
inputSchema:
  type: object
  properties:
    owner:
      type: string
    repo:
      type: string
  required: [repo]
```

### M7：事件监控（6 个 Tool）

```yaml
# 1. list_monitors
name: list_monitors
description: "列出所有监控规则及其状态。"

# 2. add_monitor
name: add_monitor
description: "添加监控规则。"
inputSchema:
  type: object
  properties:
    name:
      type: string
      description: "规则名称（唯一标识）"
    type:
      type: string
      enum: [polling, webhook]
      default: polling
    repos:
      type: array
      items:
        type: object
        properties:
          owner:
            type: string
          repo:
            type: string
    filters:
      type: array
      items:
        type: object
        properties:
          type:
            type: string
            enum: [issue_comment, issue_closed, issue_opened, new_release,
                   security_advisory, ci_failure, pr_review_requested, pr_merged]
          condition:
            type: string
            description: "可选的过滤条件表达式"
    interval:
      type: string
      description: "轮询间隔（如 '5m', '1h'）"
    notify:
      type: object
      properties:
        channels:
          type: array
          items:
            type: string
  required: [name, repos]

# 3. remove_monitor
name: remove_monitor
inputSchema:
  properties:
    name: { type: string }
  required: [name]

# 4. pause_monitor
name: pause_monitor
inputSchema:
  properties:
    name: { type: string }
  required: [name]

# 5. resume_monitor
name: resume_monitor
inputSchema:
  properties:
    name: { type: string }
  required: [name]

# 6. get_notifications
name: get_notifications
description: "获取监控系统捕获的通知。"
inputSchema:
  type: object
  properties:
    since:
      type: string
      description: "只返回该时间之后的通知"
    unread_only:
      type: boolean
      default: true
    limit:
      type: integer
      default: 20
```

## 5.3 Tool 数量汇总

| 模块 | Tool 数量 | 优先级 |
|------|-----------|--------|
| M1 仓库管理 | 8 | P0 |
| M2 Issue 管理 | 10 | P0 |
| M3 PR & Review | 7 | P0 |
| M4 Code Intelligence | 8 | P1 |
| M5 Release 管理 | 5 | P0 |
| M6 CI/CD 管理 | 6 | P1 |
| M7 事件监控 | 6 | P1 |
| M8 通知系统 | 4 | P2 |
| **合计** | **54** | — |
