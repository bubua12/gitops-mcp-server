# GitOps MCP Server

> 面向开发者和 DevOps 工程师的全功能 GitHub 运维中枢，通过 MCP 协议让 AI Agent 具备完整的 Git 平台操作能力。

## 功能特性

### Phase 1（当前版本）

- ✅ **仓库管理**：获取仓库详情、列出仓库、搜索仓库、浏览目录结构、读取文件、搜索代码
- ✅ **Issue 管理**：列出 Issues、获取 Issue 详情、搜索 Issues
- ✅ **GitHub 连接**：Token 认证、速率限制查询、健康检查

### 后续规划

- Phase 2：Issue 写操作、PR 管理、Release 管理
- Phase 3：Code Intelligence、CI/CD 管理
- Phase 4：事件监控、通知系统

## 快速开始

### 前置条件

- Go 1.21+
- GitHub Personal Access Token

### 安装

```bash
go install gitops-mcp-server/cmd/gitops-mcp@latest
```

### 配置

设置环境变量：

```bash
export GITHUB_TOKEN="your_github_token"
export GITHUB_DEFAULT_OWNER="your_username"  # 可选
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
        "GITHUB_TOKEN": "your_github_token",
        "GITHUB_DEFAULT_OWNER": "your_username"
      }
    }
  }
}
```

### SSE 模式

```bash
export MCP_TRANSPORT=sse
export MCP_PORT=18080
export GITHUB_TOKEN="your_github_token"
gitops-mcp
```

## 可用 Tools

| Tool | 描述 |
|------|------|
| `health_check` | 验证 GitHub 连接状态 |
| `get_repository` | 获取仓库详情 |
| `list_repositories` | 列出用户仓库 |
| `search_repositories` | 搜索仓库 |
| `get_repo_structure` | 获取目录结构 |
| `read_file` | 读取文件内容 |
| `search_code` | 搜索代码 |
| `list_issues` | 列出 Issues |
| `get_issue` | 获取 Issue 详情 |
| `search_issues` | 搜索 Issues |

## 项目结构

```
gitops-mcp-server/
├── cmd/gitops-mcp/         # 入口
├── internal/
│   ├── config/             # 配置系统
│   ├── github/             # GitHub API 客户端
│   └── tools/              # MCP Tools
│       ├── repo/           # 仓库管理
│       └── issue/          # Issue 管理
├── docs/                   # 产品设计文档
├── go.mod
└── README.md
```

## 开发

```bash
# 构建
go build -o gitops-mcp ./cmd/gitops-mcp

# 测试
go test ./...

# 运行（stdio 模式）
GITHUB_TOKEN=xxx ./gitops-mcp
```

## 文档

- [产品设计文档](docs/README.md)
- [API 规范](docs/05-api-spec.md)
- [架构设计](docs/04-architecture.md)

## License

MIT
