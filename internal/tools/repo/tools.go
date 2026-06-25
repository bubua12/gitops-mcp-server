package repo

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterRepoTools 注册仓库管理相关 Tools
func RegisterRepoTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	repoSvc := ghClient.Repositories()

	// get_repository
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_repository",
		Description: "获取指定 GitHub 仓库的详细信息，包括描述、语言、star/fork 数、默认分支等",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者（可选，默认使用配置中的 default_owner）",
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

		info, err := repoSvc.Get(ctx, owner, repo)
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(github.FormatRepoInfo(info)), nil
	})

	// list_repositories
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_repositories",
		Description: "列出当前认证用户的仓库列表",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"owner": map[string]any{
					"type":        "string",
					"description": "用户/组织名（可选）",
				},
				"type": map[string]any{
					"type":        "string",
					"enum":        []string{"all", "owner", "member", "public", "private"},
					"description": "仓库类型过滤，默认 all",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"created", "updated", "pushed", "full_name"},
					"description": "排序方式",
				},
				"per_page": map[string]any{
					"type":        "integer",
					"description": "每页数量，默认 30，最大 100",
				},
			},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		sort := request.GetString("sort", "full_name")
		perPage := request.GetInt("per_page", 30)

		repos, err := repoSvc.List(ctx, owner, &gogithub.RepositoryListOptions{
			Sort: sort,
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		})
		if err != nil {
			return errorResult(err.Error()), nil
		}
		return textResult(formatRepoList(repos)), nil
	})

	// search_repositories
	mcpServer.AddTool(mcp.Tool{
		Name:        "search_repositories",
		Description: "搜索 GitHub 仓库。支持按关键词、语言、star 数等条件搜索",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "搜索关键词，支持 GitHub 搜索语法，如 'language:go stars:>1000'",
				},
				"sort": map[string]any{
					"type":        "string",
					"enum":        []string{"stars", "forks", "help-wanted-issues", "updated"},
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

		repos, total, err := repoSvc.Search(ctx, query, opts)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := formatRepoList(repos)
		result += fmt.Sprintf("\n共 %d 个结果\n", total)
		return textResult(result), nil
	})

	// get_repo_structure
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_repo_structure",
		Description: "获取仓库的目录结构树。支持指定分支/tag 和子目录路径",
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
				"ref": map[string]any{
					"type":        "string",
					"description": "分支名、tag 或 commit SHA（默认：默认分支）",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "子目录路径（默认：根目录）",
				},
			},
			Required: []string{"repo"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		ref := request.GetString("ref", "")
		path := request.GetString("path", "")

		// 如果没有指定 ref，先获取默认分支
		if ref == "" {
			info, err := repoSvc.Get(ctx, owner, repo)
			if err != nil {
				return errorResult(err.Error()), nil
			}
			ref = info.DefaultBranch
		}

		// 获取目录内容
		_, _, dirContent, err := repoSvc.GetContent(ctx, owner, repo, path, ref)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		if dirContent != nil {
			return textResult(formatDirContent(dirContent, path)), nil
		}

		return textResult(fmt.Sprintf("路径 '%s' 是一个文件，请使用 read_file 工具读取", path)), nil
	})

	// read_file
	mcpServer.AddTool(mcp.Tool{
		Name:        "read_file",
		Description: "读取仓库中的文件内容。支持指定行范围",
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
				"path": map[string]any{
					"type":        "string",
					"description": "文件路径",
				},
				"ref": map[string]any{
					"type":        "string",
					"description": "分支/tag/SHA",
				},
				"line_start": map[string]any{
					"type":        "integer",
					"description": "起始行号（从 1 开始）",
				},
				"line_end": map[string]any{
					"type":        "integer",
					"description": "结束行号",
				},
			},
			Required: []string{"repo", "path"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		path := request.GetString("path", "")
		ref := request.GetString("ref", "")
		lineStart := request.GetInt("line_start", 0)
		lineEnd := request.GetInt("line_end", 0)

		content, sha, _, err := repoSvc.GetContent(ctx, owner, repo, path, ref)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := fmt.Sprintf("📄 %s (SHA: %s)\n\n", path, sha)
		if lineStart > 0 || lineEnd > 0 {
			lines := splitLines(content)
			if lineStart < 1 {
				lineStart = 1
			}
			if lineEnd < 1 || lineEnd > len(lines) {
				lineEnd = len(lines)
			}
			content = joinLines(lines[lineStart-1 : lineEnd])
			result += fmt.Sprintf("行 %d-%d:\n", lineStart, lineEnd)
		}
		result += "```\n" + content + "\n```"
		return textResult(result), nil
	})

	// search_code
	mcpServer.AddTool(mcp.Tool{
		Name:        "search_code",
		Description: "在仓库中搜索代码内容",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "搜索关键词",
				},
				"owner": map[string]any{
					"type":        "string",
					"description": "仓库所有者",
				},
				"repo": map[string]any{
					"type":        "string",
					"description": "仓库名称",
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
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		perPage := request.GetInt("per_page", 30)

		results, total, err := repoSvc.SearchCode(ctx, query, owner, repo, &gogithub.SearchOptions{
			ListOptions: gogithub.ListOptions{PerPage: perPage},
		})
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := fmt.Sprintf("🔍 代码搜索结果 (共 %d 个)\n\n", total)
		for _, r := range results {
			result += fmt.Sprintf("- **%s** / %s\n  %s\n\n", r.RepoName, r.Path, r.URL)
		}
		return textResult(result), nil
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

func formatRepoList(repos []*github.RepoInfo) string {
	if len(repos) == 0 {
		return "无仓库\n"
	}
	var result string
	for _, r := range repos {
		archived := ""
		if r.Archived {
			archived = " [已归档]"
		}
		result += fmt.Sprintf("- **%s** ⭐%d 🍴%d%s\n  %s\n",
			r.FullName, r.Stars, r.Forks, archived, r.Description)
	}
	return result
}

func formatDirContent(contents []*gogithub.RepositoryContent, path string) string {
	var result string
	if path == "" {
		path = "/"
	}
	result = fmt.Sprintf("📁 %s\n\n", path)
	for _, c := range contents {
		if c.GetType() == "dir" {
			result += fmt.Sprintf("📁 %s/\n", c.GetName())
		} else {
			result += fmt.Sprintf("📄 %s (%d bytes)\n", c.GetName(), c.GetSize())
		}
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}
