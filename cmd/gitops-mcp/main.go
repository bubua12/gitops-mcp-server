package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitops-mcp-server/internal/config"
	"gitops-mcp-server/internal/github"
	"gitops-mcp-server/internal/monitor"
	"gitops-mcp-server/internal/notify"
	"gitops-mcp-server/internal/tools/cicd"
	"gitops-mcp-server/internal/tools/intelligence"
	"gitops-mcp-server/internal/tools/issue"
	monitorTools "gitops-mcp-server/internal/tools/monitor"
	notifyTools "gitops-mcp-server/internal/tools/notify"
	"gitops-mcp-server/internal/tools/pr"
	"gitops-mcp-server/internal/tools/release"
	"gitops-mcp-server/internal/tools/repo"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "", "配置文件路径（默认自动查找 configs/config.yaml）")
	flag.Parse()

	// 加载配置（优先级：环境变量 > 配置文件 > 默认值）
	cfg, err := config.Load(config.ResolveConfigPath(*configPath))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 设置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// 创建 GitHub 客户端
	ghClient, err := github.NewClient(cfg.GitHub.Token, cfg.GitHub.DefaultOwner)
	if err != nil {
		log.Fatalf("Failed to create GitHub client: %v", err)
	}

	// 验证 GitHub 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ghClient.Ping(ctx); err != nil {
		log.Printf("Warning: GitHub connection test failed: %v", err)
	} else {
		log.Println("GitHub connection successful")
	}

	// 创建通知管理器
	notifyMgr := notify.NewManager()

	// 注册终端通知渠道（默认启用）
	notifyMgr.Register(notify.NewTerminalChannel("terminal"))

	// 创建监控引擎
	monitorEngine := monitor.NewEngine(ghClient, notifyMgr)

	// 启动监控引擎
	monitorEngine.Start(context.Background())
	defer monitorEngine.Stop()

	// 创建 MCP Server
	mcpServer := server.NewMCPServer(
		"gitops-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// 注册 health_check
	mcpServer.AddTool(mcp.Tool{
		Name:        "health_check",
		Description: "验证 GitHub 连接状态和 Token 权限",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := ghClient.Ping(ctx); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{Type: "text", Text: fmt.Sprintf("❌ GitHub connection failed: %s", err.Error())},
				},
				IsError: true,
			}, nil
		}

		remaining, resetAt, warn := ghClient.CheckRateLimit()
		result := "✅ GitHub connection is healthy\n\n"
		result += fmt.Sprintf("- API Rate Limit Remaining: %d\n", remaining)
		result += fmt.Sprintf("- Rate Limit Reset At: %s\n", resetAt.Format(time.RFC3339))
		if warn {
			result += "\n⚠️ Warning: API rate limit is getting low\n"
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: result},
			},
		}, nil
	})

	// 注册仓库管理 Tools
	repo.RegisterRepoTools(mcpServer, ghClient)

	// 注册 Issue Tools
	issue.RegisterIssueTools(mcpServer, ghClient)

	// 注册 PR Tools
	pr.RegisterPRTools(mcpServer, ghClient)

	// 注册 Release Tools
	release.RegisterReleaseTools(mcpServer, ghClient)

	// 注册 Code Intelligence Tools
	intelligence.RegisterIntelligenceTools(mcpServer, ghClient)

	// 注册 CI/CD Tools
	cicd.RegisterCICDTools(mcpServer, ghClient)

	// 注册监控 Tools
	monitorTools.RegisterMonitorTools(mcpServer, monitorEngine)

	// 注册通知 Tools
	notifyTools.RegisterNotifyTools(mcpServer, notifyMgr)

	// 优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down...")
		monitorEngine.Stop()
		os.Exit(0)
	}()

	// 启动 MCP Server
	transport := cfg.Server.Transport
	if transport == "" {
		transport = "stdio"
	}

	switch transport {
	case "stdio":
		log.Println("Starting gitops-mcp-server on stdio transport")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case "sse":
		port := cfg.Server.Port
		if port == "" {
			port = "18080"
		}
		log.Printf("Starting gitops-mcp-server on SSE transport, port: %s", port)
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(":" + port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	default:
		log.Fatalf("Unknown transport: %s", transport)
	}
}
