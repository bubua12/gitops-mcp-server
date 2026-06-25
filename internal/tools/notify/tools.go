package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/notify"
)

// RegisterNotifyTools 注册通知相关 Tools
func RegisterNotifyTools(mcpServer *server.MCPServer, mgr *notify.Manager) {
	// list_notification_channels
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_notification_channels",
		Description: "列出所有通知渠道",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channels := mgr.ListChannels()

		result := "📢 通知渠道列表\n\n"
		if len(channels) == 0 {
			result += "无通知渠道\n"
		}
		for _, ch := range channels {
			result += fmt.Sprintf("- **%s** (类型: %s)\n", ch.Name, ch.Type)
		}

		return textResult(result), nil
	})

	// send_notification
	mcpServer.AddTool(mcp.Tool{
		Name:        "send_notification",
		Description: "手动发送通知到指定渠道",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"channel": map[string]any{
					"type":        "string",
					"description": "通知渠道名称",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "通知标题",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "通知内容",
				},
				"level": map[string]any{
					"type":        "string",
					"enum":        []string{"info", "warn", "error"},
					"description": "通知级别，默认 info",
				},
			},
			Required: []string{"channel", "title"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channel := request.GetString("channel", "")
		title := request.GetString("title", "")
		body := request.GetString("body", "")
		level := request.GetString("level", "info")

		msg := notify.Notification{
			ID:        fmt.Sprintf("manual-%d", time.Now().Unix()),
			Title:     title,
			Body:      body,
			Level:     level,
			Timestamp: time.Now(),
		}

		if err := mgr.Send(ctx, channel, msg); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 通知已发送到 '%s'", channel)), nil
	})

	// test_notification
	mcpServer.AddTool(mcp.Tool{
		Name:        "test_notification",
		Description: "测试通知渠道是否正常工作",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"channel": map[string]any{
					"type":        "string",
					"description": "通知渠道名称",
				},
			},
			Required: []string{"channel"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		channel := request.GetString("channel", "")

		if err := mgr.TestChannel(ctx, channel); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 渠道 '%s' 测试成功", channel)), nil
	})
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: text},
		},
	}
}

func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: "❌ " + msg},
		},
		IsError: true,
	}
}
