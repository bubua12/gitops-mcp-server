# 二、用户画像与核心场景

## 2.1 目标用户

### 用户画像 A：独立开发者 / 技术负责人

```
👤 张工 — 资深后端工程师，维护 5+ 个开源项目
   痛点：
   · 每天在 GitHub 浏览器和终端之间切换 50+ 次
   · Issue 处理效率低，经常忘记回复别人提的 issue
   · Release 发布流程繁琐，容易遗漏 changelog
   · 想了解依赖库的最新动态但没时间跟

   期望：
   → "我只在终端里和 AI 对话，就能管好所有仓库"
```

### 用户画像 B：DevOps / SRE 工程师

```
👤 李工 — DevOps 工程师，负责 10+ 微服务的 CI/CD
   痛点：
   · CI 流水线失败需要手动排查和重试
   · 多仓库的 tag/release 管理混乱
   · 代码审查耗时，重复性问题多
   · 想监控上游依赖的 issue 和安全公告

   期望：
   → "AI 帮我巡检 CI 状态，自动重试失败的 job，有问题通知我"
```

### 用户画像 C：开源社区贡献者

```
👤 王同学 — 开源爱好者，活跃在多个社区
   痛点：
   · 提了 issue 后经常忘看回复
   · 贡献代码时不了解项目的完整上下文
   · fork 多了管理混乱，不知道哪些 PR 在等 review

   期望：
   → "帮我监控我参与的所有项目的动态，有回复就告诉我"
```

## 2.2 核心使用场景

### 场景 1：智能 Issue 全生命周期管理

```
场景描述：AI 驱动的 Issue 从创建到关闭的全流程自动化

用户说 → "帮我看看 kubernetes-manifests 仓库最近的 commit，
         如果发现潜在问题，自动创建 issue"

AI 执行流程：
  1. 调用 get_recent_commits() 获取最近 N 次提交
  2. 调用 read_file() 逐个分析变更文件
  3. 识别出潜在问题（如硬编码密码、缺失资源限制等）
  4. 调用 create_issue() 创建 Issue，自动打标签、指派
  5. 如果用户后续修复了，commit message 加 Fixes #xxx
  6. PR 合并后自动关闭 Issue

涉及 Tools：get_recent_commits, read_file, search_code,
            create_issue, add_labels, close_issue
```

### 场景 2：一键发布 Release

```
场景描述：从打 Tag 到生成 Release Notes 全自动

用户说 → "帮我给 clickhouse-operator 打一个 v1.2.0 的 tag，
         从上次 release 到现在的变更自动生成 release notes"

AI 执行流程：
  1. 调用 get_tags() 确认最新 tag 是 v1.1.0
  2. 调用 get_commits_since("v1.1.0") 获取所有新 commit
  3. 分析 commit messages，按 conventional commits 分类
     · feat: 新增 xx 功能
     · fix: 修复 xx 问题
     · docs: 文档更新
  4. 调用 create_tag("v1.2.0")
  5. 调用 create_release("v1.2.0", body=生成的release notes)
  6. 确认发布成功

涉及 Tools：get_tags, get_commits_between, create_tag,
            create_release
```

### 场景 3：Issue 反馈监控与通知

```
场景描述：监控自己和关注项目的 Issue/PR 动态

用户说 → "帮我监控这些项目：kubernetes/kubernetes,
         prometheus/prometheus, 我提的 issue 有回复就通知我"

AI 执行流程（定时任务）：
  1. 定时轮询 用户在各仓库的 issue 评论
  2. 对比上次检查时间，筛选新评论
  3. 生成摘要通知：
     · kubernetes/kubernetes#12345: @someone 回复了你的评论
     · prometheus/prometheus#678: issue 被关闭，原因：won't fix
  4. 推送到配置的通知渠道（终端/邮件/IM）

涉及 Tools：list_my_issues, get_issue_comments, 
            get_notifications, check_updates
```

### 场景 4：代码审查助手

```
场景描述：AI 自动审查 PR 代码并提 issue

用户说 → "帮我审查 open PR #23 的代码，发现问题直接提 comment"

AI 执行流程：
  1. 调用 get_pr_diff("#23") 获取代码变更
  2. 逐文件分析：
     · 安全问题（硬编码密钥、SQL 注入风险）
     · 性能问题（N+1 查询、缺失索引）
     · 代码规范（命名、注释、错误处理）
  3. 调用 create_review("#23", comments) 提交审查意见
  4. 严重问题自动创建跟踪 Issue

涉及 Tools：get_pr, get_pr_files, get_pr_diff,
            create_review, create_issue
```

### 场景 5：多仓库健康度巡检

```
场景描述：一键了解所有仓库的状态

用户说 → "帮我巡检一下我所有的仓库，有什么需要关注的"

AI 执行流程：
  1. 调用 list_repositories() 获取用户所有仓库
  2. 对每个仓库并行检查：
     · 未关闭的 Issue 数量和状态
     · 等待 review 的 PR
     · CI/CD 状态
     · 最近一次 release 的时间
     · 依赖的安全漏洞（Dependabot alerts）
  3. 生成健康度报告：
     ┌─────────────────────────────────────────┐
     │ 📊 仓库健康度巡检报告                     │
     ├─────────────────┬──────┬───────┬─────────┤
     │ 仓库            │ Issues │ PRs │ CI 状态  │
     ├─────────────────┼──────┼───────┼─────────┤
     │ k8s-manifests   │ 3 🟡  │ 1 🟡 │ ✅ 通过  │
     │ clickhouse-op   │ 0 🟢  │ 0 🟢 │ ✅ 通过  │
     │ gitops-mcp      │ 7 🔴  │ 2 🟡 │ ❌ 失败  │
     └─────────────────┴──────┴───────┴─────────┘

涉及 Tools：list_repositories, get_repo_stats, 
            list_issues, list_pull_requests, get_workflow_runs
```

### 场景 6：上游依赖监控

```
场景描述：监控你依赖的开源项目的动态

用户说 → "帮我关注 clickhouse/clickhouse 这个项目，
         有新的 release 或安全相关的 issue 通知我"

AI 执行流程（定时任务）：
  1. 调用 get_latest_release("clickhouse/clickhouse")
  2. 检查是否有新的安全相关 issue（标签含 security/CVE）
  3. 对比上次检查，有变化则通知
  4. 通知内容：
     · 🔴 安全公告：clickhouse/clickhouse#56789 
       "CVE-2026-xxxx: 远程代码执行漏洞"
     · 🆕 新版本发布：v24.3.1（包含安全修复）

涉及 Tools：get_latest_release, search_issues, 
            get_notifications
```
