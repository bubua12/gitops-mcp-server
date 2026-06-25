package monitor

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/monitor"
)

// RegisterMonitorTools 注册监控相关 Tools
func RegisterMonitorTools(mcpServer *server.MCPServer, engine *monitor.Engine) {
	// list_monitors
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_monitors",
		Description: "列出所有监控规则及其状态",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]any{},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		monitors := engine.ListMonitors()
		return textResult(monitor.FormatMonitorList(monitors)), nil
	})

	// add_monitor
	mcpServer.AddTool(mcp.Tool{
		Name:        "add_monitor",
		Description: "添加监控规则",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "规则名称（唯一标识）",
				},
				"repos": map[string]any{
					"type":        "array",
					"description": "监控仓库列表，格式: [{\"owner\":\"xxx\",\"repo\":\"xxx\"}]",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"owner": map[string]any{"type": "string"},
							"repo":  map[string]any{"type": "string"},
						},
					},
				},
				"filters": map[string]any{
					"type":        "array",
					"description": "过滤器列表，可选类型: issue_comment, new_release, ci_failure",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type":        "string",
								"enum":        []string{"issue_comment", "issue_closed", "new_release", "ci_failure", "pr_review_requested"},
								"description": "事件类型",
							},
						},
					},
				},
				"interval": map[string]any{
					"type":        "string",
					"description": "轮询间隔，如 '5m', '1h'，默认 5m",
				},
				"notify_channels": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "通知渠道名称列表",
				},
			},
			Required: []string{"name", "repos"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")
		intervalStr := request.GetString("interval", "5m")

		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return errorResult(fmt.Sprintf("invalid interval: %s", intervalStr)), nil
		}

		// 解析 repos
		repos := parseRepos(request)

		// 解析 filters
		filters := parseFilters(request)

		// 解析 notify_channels
		channels := getStringArray(request, "notify_channels")

		m := &monitor.Monitor{
			Name:     name,
			Repos:    repos,
			Filters:  filters,
			Interval: interval,
			NotifyTo: channels,
			Running:  true,
		}

		if err := engine.AddMonitor(m); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已添加监控规则: %s", name)), nil
	})

	// remove_monitor
	mcpServer.AddTool(mcp.Tool{
		Name:        "remove_monitor",
		Description: "移除监控规则",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "规则名称",
				},
			},
			Required: []string{"name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")

		if err := engine.RemoveMonitor(name); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已移除监控规则: %s", name)), nil
	})

	// pause_monitor
	mcpServer.AddTool(mcp.Tool{
		Name:        "pause_monitor",
		Description: "暂停监控规则",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "规则名称",
				},
			},
			Required: []string{"name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")

		if err := engine.PauseMonitor(name); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已暂停监控规则: %s", name)), nil
	})

	// resume_monitor
	mcpServer.AddTool(mcp.Tool{
		Name:        "resume_monitor",
		Description: "恢复监控规则",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "规则名称",
				},
			},
			Required: []string{"name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := request.GetString("name", "")

		if err := engine.ResumeMonitor(name); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已恢复监控规则: %s", name)), nil
	})

	// get_events
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_events",
		Description: "获取监控系统捕获的事件日志",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"since": map[string]any{
					"type":        "string",
					"description": "只返回该时间之后的事件（ISO 8601 格式）",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "返回数量限制，默认 20",
				},
			},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sinceStr := request.GetString("since", "")
		limit := request.GetInt("limit", 20)

		since := time.Time{}
		if sinceStr != "" {
			t, err := time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				return errorResult(fmt.Sprintf("invalid since time: %s", sinceStr)), nil
			}
			since = t
		}

		events := engine.GetEvents(since, limit)
		return textResult(monitor.FormatEventList(events)), nil
	})
}

func parseRepos(request mcp.CallToolRequest) []monitor.RepoRef {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := args["repos"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}

	var repos []monitor.RepoRef
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		owner, _ := obj["owner"].(string)
		repo, _ := obj["repo"].(string)
		if owner != "" && repo != "" {
			repos = append(repos, monitor.RepoRef{Owner: owner, Repo: repo})
		}
	}
	return repos
}

func parseFilters(request mcp.CallToolRequest) []monitor.Filter {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := args["filters"]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}

	var filters []monitor.Filter
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		filterType, _ := obj["type"].(string)
		if filterType != "" {
			filters = append(filters, monitor.Filter{Type: filterType})
		}
	}
	return filters
}

func getStringArray(request mcp.CallToolRequest, key string) []string {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := args[key]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	var result []string
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
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
