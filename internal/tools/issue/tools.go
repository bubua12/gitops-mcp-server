package issue

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterIssueTools 注册 Issue 相关 Tools（读操作）
func RegisterIssueTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	issueSvc := ghClient.Issues()

	// list_issues
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_issues",
		Description: "列出仓库的 Issues。支持按状态、标签、指派人过滤",
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
				"state": map[string]any{
					"type":        "string",
					"enum":        []string{"open", "closed", "all"},
					"description": "Issue 状态，默认 open",
				},
				"labels": map[string]any{
					"type":        "string",
					"description": "逗号分隔的标签名",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "comments"},
					"description": "排序方式",
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
		state := request.GetString("state", "open")
		labels := request.GetString("labels", "")
		sort := request.GetString("sort", "created")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.IssueListByRepoOptions{
			State:     state,
			Sort:      sort,
			Direction: "desc",
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if labels != "" {
			opts.Labels = []string{labels}
		}

		issues, err := issueSvc.List(ctx, owner, repo, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatIssueList(issues, owner, repo)), nil
	})

	// get_issue
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_issue",
		Description: "获取 Issue 详情，包含全部评论",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		issue, err := issueSvc.Get(ctx, owner, repo, number)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := github.FormatIssue(issue)

		// 获取评论
		comments, err := issueSvc.ListComments(ctx, owner, repo, number, nil)
		if err == nil && len(comments) > 0 {
			result += "\n" + github.FormatCommentList(comments, number)
		}

		return textResult(result), nil
	})

	// search_issues
	mcpServer.AddTool(mcp.Tool{
		Name:        "search_issues",
		Description: "使用 GitHub 搜索语法搜索 Issue",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "搜索查询，支持 GitHub Issue 搜索语法，如 'is:issue is:open label:bug'",
				},
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者（可选）",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称（可选）",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "comments"},
					"description": "排序方式",
				},
				"per_page": map[string]any{
					"type":        "integer",
					"description": "每页数量，默认 30",
				},
			},
			Required: []string{"query"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := request.GetString("query", "")
		sort := request.GetString("sort", "")
		perPage := request.GetInt("per_page", 30)

		opts := &gogithub.SearchOptions{
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		}
		if sort != "" {
			opts.Sort = sort
		}

		issues, total, err := issueSvc.Search(ctx, query, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var result string
		result = fmt.Sprintf("🔍 Issue 搜索结果 (共 %d 个)\n\n", total)
		for _, issue := range issues {
			state := "🟢"
			if issue.State == "closed" {
				state = "🔴"
			}
			if issue.Repo != "" {
				result += fmt.Sprintf("%s **%s** #%d %s (by %s)\n", state, issue.Repo, issue.Number, issue.Title, issue.User)
			} else {
				result += fmt.Sprintf("%s #%d %s (by %s)\n", state, issue.Number, issue.Title, issue.User)
			}
		}
		return textResult(result), nil
	})

	// create_issue
	mcpServer.AddTool(mcp.Tool{
		Name:        "create_issue",
		Description: "创建新 Issue。支持设置标签、指派人",
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
				"title": map[string]any{
					"type":        "string",
					"description": "Issue 标题",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Issue 描述，支持 Markdown",
				},
				"labels": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "标签列表",
				},
				"assignees": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "指派人列表",
				},
			},
			Required: []string{"repo", "title"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		title := request.GetString("title", "")
		body := request.GetString("body", "")

		req := &gogithub.IssueRequest{
			Title: &title,
			Body:  &body,
		}

		// 处理 labels
		if labels := getStringArray(request, "labels"); len(labels) > 0 {
			req.Labels = &labels
		}

		// 处理 assignees
		if assignees := getStringArray(request, "assignees"); len(assignees) > 0 {
			req.Assignees = &assignees
		}

		issue, err := issueSvc.Create(ctx, owner, repo, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已创建 Issue #%d\n\n%s", issue.Number, github.FormatIssue(issue))), nil
	})

	// update_issue
	mcpServer.AddTool(mcp.Tool{
		Name:        "update_issue",
		Description: "更新 Issue 的标题、描述、状态、标签等",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "新标题（可选）",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "新描述（可选）",
				},
				"state": map[string]any{
					"type":        "string",
					"enum":        []string{"open", "closed"},
					"description": "新状态（可选）",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)
		title := request.GetString("title", "")
		body := request.GetString("body", "")
		state := request.GetString("state", "")

		req := &gogithub.IssueRequest{}
		if title != "" {
			req.Title = &title
		}
		if body != "" {
			req.Body = &body
		}
		if state != "" {
			req.State = &state
		}

		issue, err := issueSvc.Update(ctx, owner, repo, number, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ Issue #%d 已更新\n\n%s", issue.Number, github.FormatIssue(issue))), nil
	})

	// close_issue
	mcpServer.AddTool(mcp.Tool{
		Name:        "close_issue",
		Description: "关闭 Issue。支持指定关闭原因",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
				"reason": map[string]any{
					"type":        "string",
					"enum":        []string{"completed", "not_planned", "duplicate"},
					"description": "关闭原因，默认 completed",
				},
			},
			Required: []string{"repo", "number"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)
		reason := request.GetString("reason", "completed")

		issue, err := issueSvc.Close(ctx, owner, repo, number, reason)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ Issue #%d 已关闭 (原因: %s)\n\n%s", issue.Number, reason, github.FormatIssue(issue))), nil
	})

	// add_comment
	mcpServer.AddTool(mcp.Tool{
		Name:        "add_comment",
		Description: "对 Issue 添加评论",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "评论内容，支持 Markdown",
				},
			},
			Required: []string{"repo", "number", "body"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)
		body := request.GetString("body", "")

		comment, err := issueSvc.AddComment(ctx, owner, repo, number, body)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 评论已添加 (ID: %d)\n\n%s", comment.ID, comment.Body)), nil
	})

	// add_labels
	mcpServer.AddTool(mcp.Tool{
		Name:        "add_labels",
		Description: "为 Issue 添加标签",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
				"labels": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "要添加的标签列表",
				},
			},
			Required: []string{"repo", "number", "labels"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		labels := getStringArray(request, "labels")

		if err := issueSvc.AddLabels(ctx, owner, repo, number, labels); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已为 Issue #%d 添加标签: %v", number, labels)), nil
	})

	// remove_labels
	mcpServer.AddTool(mcp.Tool{
		Name:        "remove_labels",
		Description: "移除 Issue 的标签",
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
				"number": map[string]any{
					"type":        "integer",
					"description": "Issue 编号",
				},
				"labels": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "要移除的标签列表",
				},
			},
			Required: []string{"repo", "number", "labels"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		number := request.GetInt("number", 0)

		labels := getStringArray(request, "labels")

		if err := issueSvc.RemoveLabels(ctx, owner, repo, number, labels); err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已从 Issue #%d 移除标签: %v", number, labels)), nil
	})
}

// getStringArray 从请求中获取字符串数组参数
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
