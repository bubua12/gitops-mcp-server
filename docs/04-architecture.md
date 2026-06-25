# 四、系统架构设计

## 4.1 技术选型

| 维度 | 选型 | 理由 |
|------|------|------|
| **开发语言** | Go | 项目已初始化 go.mod；Go 生态的 GitHub SDK 成熟（google/go-github）；单二进制部署简单 |
| **MCP SDK** | mark3labs/mcp-go | Go 生态最活跃的 MCP SDK，支持 stdio/SSE/Streamable HTTP |
| **GitHub 客户端** | google/go-github v68+ | Google 官方维护，API 覆盖最全，GraphQL 也可选用 shurcooL/githubv4 |
| **配置管理** | Viper | 支持 YAML、环境变量、远程配置，生态成熟 |
| **日志** | slog（标准库） | Go 1.21+ 内置，结构化日志，零外部依赖 |
| **缓存** | 内存 + 可选 Redis | 初期纯内存（go-cache），后续可插拔 Redis |
| **测试** | testify + httptest | GitHub API mock 测试 |

## 4.2 整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                        MCP 客户端层                                  │
│              Claude Code / Cline / VS Code / 自定义客户端             │
└──────────────────────┬──────────────────────────────────────────────┘
                       │ MCP Protocol (stdio / SSE / HTTP)
                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     GitOps MCP Server                                │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    MCP Transport Layer                       │   │
│  │            stdio │ SSE │ Streamable HTTP                     │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │                                      │
│  ┌──────────────────────────▼──────────────────────────────────┐   │
│  │                    Tool Router / Dispatcher                  │   │
│  │         根据 tool name 分发到对应的 Handler                    │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │                                      │
│  ┌──────────────────────────▼──────────────────────────────────┐   │
│  │                    Tool Handlers (业务层)                     │   │
│  │                                                              │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐    │   │
│  │  │ Repo   │ │ Issue  │ │ PR &   │ │ Code   │ │ Release│    │   │
│  │  │ Handler│ │ Handler│ │ Review │ │ Intel  │ │ Handler│    │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘    │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐                          │   │
│  │  │ CI/CD  │ │ Monitor│ │ Notify │                          │   │
│  │  │ Handler│ │ Handler│ │ Handler│                          │   │
│  │  └────────┘ └────────┘ └────────┘                          │   │
│  └──────────────────────────┬──────────────────────────────────┘   │
│                              │                                      │
│  ┌──────────────────────────▼──────────────────────────────────┐   │
│  │                    服务层 (Services)                          │   │
│  │                                                              │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐               │   │
│  │  │ GitHub     │ │ Monitor    │ │ Notify     │               │   │
│  │  │ Service    │ │ Service    │ │ Service    │               │   │
│  │  │            │ │            │ │            │               │   │
│  │  │ · REST API │ │ · Polling  │ │ · Terminal │               │   │
│  │  │ · GraphQL  │ │ · Webhook  │ │ · Webhook  │               │   │
│  │  │ · 批量优化  │ │ · 去重过滤  │ │ · Email    │               │   │
│  │  └─────┬──────┘ └─────┬──────┘ └─────┬──────┘               │   │
│  └────────┼──────────────┼──────────────┼──────────────────────┘   │
│           │              │              │                           │
│  ┌────────▼──────────────▼──────────────▼──────────────────────┐   │
│  │                    基础设施层                                  │   │
│  │                                                              │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌──────────┐ │   │
│  │  │ Config     │ │ Cache      │ │ Logger     │ │ Rate     │ │   │
│  │  │ Manager    │ │ Manager    │ │ (slog)     │ │ Limiter  │ │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └──────────┘ │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    外部服务层                                         │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  GitHub API  │  │  GitHub      │  │  通知渠道     │              │
│  │  REST v3     │  │  Webhooks    │  │  (Webhook/   │              │
│  │  GraphQL v4  │  │              │  │   Email)     │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
└─────────────────────────────────────────────────────────────────────┘
```

## 4.3 目录结构

```
gitops-mcp-server/
├── cmd/
│   └── gitops-mcp/
│       └── main.go                  # 入口，初始化配置 + 启动 MCP Server
├── internal/
│   ├── config/
│   │   ├── config.go                # 配置结构体定义
│   │   ├── loader.go                # Viper 加载配置
│   │   └── defaults.go              # 默认值
│   ├── github/
│   │   ├── client.go                # GitHub API 客户端封装
│   │   ├── auth.go                  # 认证管理（PAT / App）
│   │   ├── rate_limiter.go          # 速率限制感知
│   │   └── types.go                 # GitHub 相关类型定义
│   ├── tools/                       # MCP Tool 实现
│   │   ├── repo/                    # M1: 仓库管理
│   │   │   ├── handler.go
│   │   │   └── tools.go             # Tool 注册定义
│   │   ├── issue/                   # M2: Issue 管理
│   │   │   ├── handler.go
│   │   │   ├── tools.go
│   │   │   └── templates.go         # Issue 模板
│   │   ├── pr/                      # M3: PR & Review
│   │   │   ├── handler.go
│   │   │   └── tools.go
│   │   ├── intelligence/            # M4: Code Intelligence
│   │   │   ├── handler.go
│   │   │   ├── analyzer.go          # 变更分析
│   │   │   └── tools.go
│   │   ├── release/                 # M5: Release 管理
│   │   │   ├── handler.go
│   │   │   ├── notes_generator.go   # Release Notes 生成器
│   │   │   └── tools.go
│   │   ├── cicd/                    # M6: CI/CD 管理
│   │   │   ├── handler.go
│   │   │   └── tools.go
│   │   ├── monitor/                 # M7: 事件监控
│   │   │   ├── handler.go
│   │   │   ├── poller.go            # Polling 引擎
│   │   │   ├── webhook.go           # Webhook Server
│   │   │   └── tools.go
│   │   └── notify/                  # M8: 通知系统
│   │       ├── handler.go
│   │       ├── channels/
│   │       │   ├── terminal.go
│   │       │   ├── webhook.go
│   │       │   └── email.go
│   │       └── tools.go
│   ├── services/                    # 业务服务层
│   │   ├── github_service.go        # GitHub API 服务（聚合 + 缓存）
│   │   ├── monitor_service.go       # 监控调度服务
│   │   └── notify_service.go        # 通知分发服务
│   └── infra/                       # 基础设施
│       ├── cache/
│       │   ├── cache.go             # 缓存接口
│       │   ├── memory.go            # 内存缓存实现
│       │   └── redis.go             # Redis 缓存实现（可选）
│       ├── logger/
│       │   └── logger.go            # 日志初始化
│       └── ratelimit/
│           └── limiter.go           # 速率限制
├── pkg/                             # 可导出的公共包
│   └── mcp/                         # MCP 工具注册和路由
│       ├── server.go                # MCP Server 初始化和启动
│       ├── router.go                # Tool 路由分发
│       └── middleware.go            # 中间件（日志、指标、错误处理）
├── configs/
│   ├── config.example.yaml          # 示例配置
│   └── config.minimal.yaml          # 最小配置
├── docs/                            # 文档
│   ├── 01-product-overview.md
│   ├── 02-user-scenarios.md
│   ├── 03-feature-modules.md
│   ├── 04-architecture.md
│   ├── 05-api-spec.md
│   ├── 06-security.md
│   └── 07-roadmap.md
├── scripts/
│   ├── install.sh                   # 安装脚本
│   └── claude-code-setup.sh         # Claude Code 配置一键脚本
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 4.4 核心流程：一次 Tool 调用的完整链路

以 `create_issue` 为例：

```
用户在 Claude Code 中说："帮我在 myrepo 创建一个 issue"
        │
        ▼
Claude Code 解析意图，调用 MCP Tool: create_issue
        │
        ▼  (MCP Protocol - stdio)
┌─ GitOps MCP Server ──────────────────────────────────┐
│                                                       │
│  1. Transport Layer 接收 JSON-RPC 请求                 │
│     └→ 解析 method: "tools/call"                      │
│     └→ 解析 tool: "create_issue"                      │
│                                                       │
│  2. Router 匹配到 issue.Handler.CreateIssue           │
│                                                       │
│  3. Middleware 执行                                    │
│     └→ 日志记录（请求参数）                             │
│     └→ 速率限制检查                                    │
│     └→ 参数校验                                        │
│                                                       │
│  4. issue.Handler.CreateIssue 执行                     │
│     └→ 检查缓存（无缓存，是写操作）                      │
│     └→ 调用 GitHub Service                             │
│         └→ 检查 Token 权限                             │
│         └→ 调用 GitHub REST API:                       │
│            POST /repos/{owner}/{repo}/issues           │
│         └→ 更新速率限制计数器                           │
│     └→ 缓存失效：该仓库的 issue 列表缓存清除             │
│     └→ 触发通知（如果配置了监控规则）                     │
│                                                       │
│  5. 构造 MCP 响应                                      │
│     └→ 返回创建的 Issue 信息（URL、number、状态）        │
│                                                       │
└───────────────────────────────────────────────────────┘
        │
        ▼  (MCP Protocol - stdio)
Claude Code 收到结果，展示给用户
        │
        ▼
"✅ 已创建 Issue #42: https://github.com/xxx/myrepo/issues/42"
```

## 4.5 并发与生命周期管理

```
┌─────────────────────────────────────────────┐
│            Server Lifecycle                  │
│                                              │
│  main()                                      │
│    ├─ 加载配置 (config.Loader)               │
│    ├─ 初始化基础设施                          │
│    │   ├─ Logger (slog)                      │
│    │   ├─ Cache (memory/redis)               │
│    │   └─ Rate Limiter                       │
│    ├─ 初始化 GitHub Client                   │
│    │   ├─ 验证 Token                         │
│    │   └─ 检查权限范围                        │
│    ├─ 初始化 Services                        │
│    │   ├─ GitHubService                      │
│    │   ├─ MonitorService                     │
│    │   └─ NotifyService                      │
│    ├─ 注册 MCP Tools (所有模块)               │
│    ├─ 启动后台任务                            │
│    │   ├─ Monitor Poller (goroutine)         │
│    │   ├─ Webhook Server (goroutine, 可选)   │
│    │   └─ 健康检查 (goroutine, 可选)         │
│    ├─ 启动 MCP Server (阻塞)                 │
│    │   ├─ stdio 模式：读 stdin，写 stdout     │
│    │   └─ SSE/HTTP 模式：监听端口             │
│    └─ 优雅退出                                │
│        ├─ 停止 Monitor Poller                │
│        ├─ 关闭 Webhook Server                │
│        ├─ 刷出缓存                            │
│        └─ 关闭 GitHub Client                 │
└─────────────────────────────────────────────┘
```

## 4.6 扩展性设计

| 扩展点 | 机制 | 说明 |
|--------|------|------|
| 新增 MCP Tool | 接口注册 | 实现 Handler 接口后在 Router 注册，零配置 |
| 新增通知渠道 | Channel 接口 | 实现 `notify.Channel` 接口即可新增渠道 |
| 新增 Git 平台 | Client 接口 | `github.Client` 抽象为接口，可替换为 GitLab/Gitea 实现 |
| 新增缓存后端 | Cache 接口 | `cache.Cache` 接口，可实现 Redis/SQLite 等 |
| 新增监控模式 | Monitor 接口 | 除 Polling/Webhook 外可扩展 SSE 等新模式 |

### 核心接口定义

```go
// internal/github/client.go
type Client interface {
    Repositories() RepositoryService
    Issues() IssueService
    PullRequests() PullRequestService
    Actions() ActionsService
    Git() GitService
    Notifications() NotificationsService
}

// internal/infra/cache/cache.go
type Cache interface {
    Get(ctx context.Context, key string) ([]byte, bool)
    Set(ctx context.Context, key string, value []byte, ttl time.Duration)
    Delete(ctx context.Context, key string)
    Flush(ctx context.Context)
}

// internal/tools/notify/channels/channel.go
type Channel interface {
    Name() string
    Type() string
    Send(ctx context.Context, msg Notification) error
    Test(ctx context.Context) error
}

// internal/tools/monitor/poller.go
type Monitor interface {
    Name() string
    Start(ctx context.Context) error
    Stop() error
    Status() MonitorStatus
}
```
