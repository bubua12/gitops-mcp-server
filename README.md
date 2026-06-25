# 🔧 GitOps MCP Server

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![MCP Protocol](https://img.shields.io/badge/MCP-1.0-FF6B35?style=flat-square&logo=modelcontextprotocol&logoColor=white)](https://modelcontextprotocol.io)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/wangxl/gitops-mcp-server?style=flat-square&logo=github)](https://github.com/wangxl/gitops-mcp-server)

> 🚀 **面向开发者和 DevOps 工程师的全功能 GitHub 运维中枢**
>
> 通过 MCP 协议让 AI Agent 具备完整的 Git 平台操作能力，实现 **"对话即运维"** 的开发体验。

---

## ✨ 核心特性

<table>
<tr>
<td width="50%">

### 📦 仓库管理
- 🔍 搜索仓库、浏览目录结构
- 📄 读取文件内容（支持行范围）
- 🔎 代码搜索

### 📋 Issue 管理
- ✏️ 创建、更新、关闭 Issue
- 💬 添加评论、管理标签
- 🔍 搜索 Issues

### 🔀 PR 管理
- 📝 创建、审查、合并 PR
- 📊 获取 diff 和变更文件
- 👀 审查历史查看

</td>
<td width="50%">

### 🏷️ Release 管理
- 🎯 创建 Tag 和 Release
- 📝 自动生成 Release Notes
- 📋 列出历史版本

### ⚙️ CI/CD 管理
- 🚀 触发、重试、取消工作流
- 📊 CI/CD 巡检报告
- 📝 查看 Job 日志

### 📡 事件监控 & 通知
- 🔔 实时监控 Issue/Release/CI 事件
- 📢 多渠道通知（Terminal/Webhook）
- 📋 事件日志回溯

</td>
</tr>
</table>

---

## 🛠️ 48 个 MCP Tools

<details>
<summary><b>📦 M1 仓库管理（7 个）</b></summary>

| Tool | 描述 |
|------|------|
| `health_check` | 验证 GitHub 连接状态和速率限制 |
| `get_repository` | 获取仓库详情 |
| `list_repositories` | 列出用户仓库 |
| `search_repositories` | 搜索仓库 |
| `get_repo_structure` | 获取目录结构树 |
| `read_file` | 读取文件内容（支持行范围） |
| `search_code` | 搜索代码内容 |

</details>

<details>
<summary><b>📋 M2 Issue 管理（9 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_issues` | 列出 Issues（支持状态/标签过滤） |
| `get_issue` | 获取 Issue 详情和评论 |
| `search_issues` | 搜索 Issues |
| `create_issue` | 创建 Issue（支持标签、指派） |
| `update_issue` | 更新 Issue |
| `close_issue` | 关闭 Issue（支持原因） |
| `add_comment` | 添加评论 |
| `add_labels` | 添加标签 |
| `remove_labels` | 移除标签 |

</details>

<details>
<summary><b>🔀 M3 PR 管理（7 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_pull_requests` | 列出 PR |
| `get_pull_request` | 获取 PR 详情 |
| `get_pr_diff` | 获取 PR diff |
| `get_pr_files` | 获取变更文件列表 |
| `create_pull_request` | 创建 PR |
| `create_review` | 提交代码审查 |
| `merge_pull_request` | 合并 PR（merge/squash/rebase） |

</details>

<details>
<summary><b>🧠 M4 Code Intelligence（4 个）</b></summary>

| Tool | 描述 |
|------|------|
| `get_recent_commits` | 获取最近提交记录 |
| `get_commit_detail` | 获取 commit 详情 |
| `compare_refs` | 对比两个 ref 差异 |
| `analyze_changes` | 分析变更影响（文件类型、目录、风险） |

</details>

<details>
<summary><b>🏷️ M5 Release 管理（6 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_tags` | 列出 Tags |
| `create_tag` | 创建 Tag |
| `list_releases` | 列出 Releases |
| `create_release` | 创建 Release |
| `get_latest_release` | 获取最新 Release |
| `generate_release_notes` | 自动生成 Release Notes |

</details>

<details>
<summary><b>⚙️ M6 CI/CD 管理（6 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_workflows` | 列出工作流 |
| `list_workflow_runs` | 列出运行记录 |
| `get_workflow_run` | 获取运行详情（含 Jobs/Steps） |
| `get_workflow_summary` | CI/CD 巡检报告 |
| `rerun_workflow` | 重新运行失败的工作流 |
| `cancel_workflow` | 取消正在运行的工作流 |

</details>

<details>
<summary><b>📡 M7 事件监控（6 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_monitors` | 列出监控规则 |
| `add_monitor` | 添加监控规则 |
| `remove_monitor` | 移除监控规则 |
| `pause_monitor` | 暂停监控 |
| `resume_monitor` | 恢复监控 |
| `get_events` | 获取事件日志 |

</details>

<details>
<summary><b>📢 M8 通知系统（3 个）</b></summary>

| Tool | 描述 |
|------|------|
| `list_notification_channels` | 列出通知渠道 |
| `send_notification` | 发送通知 |
| `test_notification` | 测试通知渠道 |

</details>

---

## 🚀 快速开始

### 前置条件

- ![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go) 
- ![GitHub](https://img.shields.io/badge/GitHub-Personal_Access_Token-181717?style=flat-square&logo=github)

### 安装

```bash
# 方式 1: go install
go install github.com/wangxl/gitops-mcp-server/cmd/gitops-mcp@latest

# 方式 2: 源码编译
git clone https://github.com/wangxl/gitops-mcp-server.git
cd gitops-mcp-server
go build -o gitops-mcp ./cmd/gitops-mcp
```

### 配置

**环境变量方式（推荐）：**

```bash
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export GITHUB_DEFAULT_OWNER="your_username"  # 可选
```

**配置文件方式：**

```bash
cp configs/config.example.yaml config.yaml
# 编辑 config.yaml
```

### Claude Code 集成

在 Claude Code 的 MCP 配置中添加：

```json
{
  "mcpServers": {
    "gitops-mcp-server": {
      "command": "gitops-mcp",
      "args": [],
      "env": {
        "GITHUB_TOKEN": "ghp_xxxxxxxxxxxx",
        "GITHUB_DEFAULT_OWNER": "your_username"
      }
    }
  }
}
```

### SSE 模式（远程访问）

```bash
export MCP_TRANSPORT=sse
export MCP_PORT=18080
export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
./gitops-mcp
```

---

## 💡 使用示例

连接到 Claude Code 后，你可以直接用自然语言操作：

```
👤 你: 帮我看一下 wangxl/kubernetes-manifests 仓库最近有什么 Issue

🤖 AI: 我来帮你查看该仓库的 Issues...
       [调用 list_issues]

👤 你: 帮我创建一个 Issue，标题是"优化部署脚本"

🤖 AI: 好的，我来创建 Issue...
       [调用 create_issue]

👤 你: 帮我给 clickhouse-operator 打一个 v1.2.0 的 tag 并发 release

🤖 AI: 我来帮你创建 Tag 和 Release...
       [调用 create_tag → generate_release_notes → create_release]

👤 你: 帮我监控 kubernetes/kubernetes 的新 release，有更新就通知我

🤖 AI: 我来设置监控规则...
       [调用 add_monitor]
```

---

## 🏗️ 项目结构

```
gitops-mcp-server/
├── cmd/gitops-mcp/
│   └── main.go                    # 🚀 入口
├── internal/
│   ├── config/                    # ⚙️ 配置系统
│   │   ├── config.go
│   │   └── loader.go
│   ├── github/                    # 🐙 GitHub API 客户端
│   │   ├── client.go
│   │   ├── repository.go
│   │   ├── issue.go
│   │   ├── pull_request.go
│   │   ├── release.go
│   │   ├── git.go
│   │   └── actions.go
│   ├── monitor/                   # 📡 事件监控引擎
│   │   └── monitor.go
│   ├── notify/                    # 📢 通知系统
│   │   └── notify.go
│   └── tools/                     # 🔧 MCP Tools 实现
│       ├── repo/
│       ├── issue/
│       ├── pr/
│       ├── release/
│       ├── intelligence/
│       ├── cicd/
│       ├── monitor/
│       └── notify/
├── configs/
│   └── config.example.yaml        # 📝 配置示例
├── docs/                          # 📚 产品设计文档
│   ├── 01-product-overview.md
│   ├── 02-user-scenarios.md
│   ├── 03-feature-modules.md
│   ├── 04-architecture.md
│   ├── 05-api-spec.md
│   ├── 06-security.md
│   └── 07-roadmap.md
├── go.mod
├── go.sum
└── README.md
```

---

## 🛠️ 开发

```bash
# 构建
go build -o gitops-mcp ./cmd/gitops-mcp

# 测试
go test ./...

# 运行（stdio 模式）
GITHUB_TOKEN=xxx ./gitops-mcp

# 运行（SSE 模式）
MCP_TRANSPORT=sse MCP_PORT=18080 GITHUB_TOKEN=xxx ./gitops-mcp
```

---

## 📚 文档

- 📖 [产品设计文档](docs/README.md)
- 🔌 [API 规范](docs/05-api-spec.md)
- 🏗️ [架构设计](docs/04-architecture.md)
- 🔒 [安全设计](docs/06-security.md)
- 🗺️ [里程碑规划](docs/07-roadmap.md)

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/xxx`)
3. 提交变更 (`git commit -m 'feat: add xxx'`)
4. 推送到分支 (`git push origin feature/xxx`)
5. 创建 Pull Request

---

## 📄 License

[MIT License](LICENSE)

---

<div align="center">

**用 ❤️ 和 Go 构建**

![Go](https://img.shields.io/badge/Powered_by-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![MCP](https://img.shields.io/badge/Protocol-MCP-FF6B35?style=for-the-badge&logo=modelcontextprotocol&logoColor=white)
![GitHub](https://img.shields.io/badge/Platform-GitHub-181717?style=for-the-badge&logo=github&logoColor=white)

</div>
