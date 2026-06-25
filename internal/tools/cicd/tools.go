package cicd

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterCICDTools 注册 CI/CD 管理相关 Tools
func RegisterCICDTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	actionsSvc := ghClient.Actions()

	// list_workflows
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_workflows",
		Description: "列出仓库的 GitHub Actions 工作流",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
			},
			Required: []string{"repo"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")

		workflows, err := actionsSvc.ListWorkflows(ctx, owner, repo)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatWorkflowList(workflows, owner, repo)), nil
	})

	// list_workflow_runs
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_workflow_runs",
		Description: "列出工作流的运行记录",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"branch": map[string]any{
					"type":        "string",
					"description": "分支过滤",
				},
				"status": map[string]any{
					"type":        "string",
					"enum":        []string{"queued", "in_progress", "completed", "success", "failure", "cancelled"},
					"description": "状态过滤",
				},
				"per_page": map[string]any{
					"type":        "integer",
					"description": "每页数量，默认 30",
				},
			},
			Required: []string{"repo"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		branch := request.GetString("branch", "")
		status := request.GetString("status", "")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.ListWorkflowRunsOptions{
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if branch != "" {
			opts.Branch = branch
		}
		if status != "" {
			opts.Status = status
		}

		runs, err := actionsSvc.ListWorkflowRuns(ctx, owner, repo, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatWorkflowRunList(runs, owner, repo)), nil
	})

	// get_workflow_run
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_workflow_run",
		Description: "获取工作流运行详情",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"run_id": map[string]any{
					"type":        "integer",
					"description": "运行 ID",
				},
			},
			Required: []string{"repo", "run_id"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		runID := int64(request.GetInt("run_id", 0))

		run, err := actionsSvc.GetWorkflowRun(ctx, owner, repo, runID)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := github.FormatWorkflowRunDetail(run)

		// 获取 Jobs
		jobs, err := actionsSvc.ListJobs(ctx, owner, repo, runID)
		if err == nil && len(jobs) > 0 {
			result += "\n" + github.FormatJobList(jobs)
		}

		return textResult(result), nil
	})

	// get_workflow_summary
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_workflow_summary",
		Description: "一次调用获取所有工作流的最近运行状态汇总。适合巡检",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
			},
			Required: []string{"repo"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")

		workflows, err := actionsSvc.ListWorkflows(ctx, owner, repo)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var result string
		result = fmt.Sprintf("## CI/CD 巡检报告: %s/%s\n\n", owner, repo)

		for _, w := range workflows {
			// 获取每个 workflow 的最近一次运行
			runs, err := actionsSvc.ListWorkflowRuns(ctx, owner, repo, &gogithub.ListWorkflowRunsOptions{
				ListOptions: gogithub.ListOptions{PerPage: 1},
			})
			if err != nil || len(runs) == 0 {
				result += fmt.Sprintf("- **%s**: 无运行记录\n", w.Name)
				continue
			}

			run := runs[0]
			icon := "⏳"
			if run.Status == "completed" {
				if run.Conclusion == "success" {
					icon = "✅"
				} else if run.Conclusion == "failure" {
					icon = "❌"
				}
			}

			result += fmt.Sprintf("%s **%s** [%s/%s] - %s\n", icon, w.Name, run.Status, run.Conclusion, run.HTMLURL)
		}

		return textResult(result), nil
	})

	// rerun_workflow
	mcpServer.AddTool(mcp.Tool{
		Name:        "rerun_workflow",
		Description: "重新运行失败的工作流",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"run_id": map[string]any{
					"type":        "integer",
					"description": "运行 ID",
				},
			},
			Required: []string{"repo", "run_id"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		runID := int64(request.GetInt("run_id", 0))

		if err := actionsSvc.RerunWorkflow(ctx, owner, repo, runID); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已触发重新运行: Run #%d", runID)), nil
	})

	// cancel_workflow
	mcpServer.AddTool(mcp.Tool{
		Name:        "cancel_workflow",
		Description: "取消正在运行的工作流",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
				},
				"run_id": map[string]any{
					"type":        "integer",
					"description": "运行 ID",
				},
			},
			Required: []string{"repo", "run_id"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		runID := int64(request.GetInt("run_id", 0))

		if err := actionsSvc.CancelWorkflow(ctx, owner, repo, runID); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已取消运行: Run #%d", runID)), nil
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
