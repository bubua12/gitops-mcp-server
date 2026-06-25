package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gitops-mcp-server/internal/config"
	"gitops-mcp-server/internal/github"
	"gitops-mcp-server/internal/tools/issue"
	"gitops-mcp-server/internal/tools/repo"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
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
