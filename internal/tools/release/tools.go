package release

import (
	"context"
	"fmt"

	gogithub "github.com/google/go-github/v68/github"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitops-mcp-server/internal/github"
)

// RegisterReleaseTools 注册 Release 管理相关 Tools
func RegisterReleaseTools(mcpServer *server.MCPServer, ghClient *github.Client) {
	releaseSvc := ghClient.Releases()
	gitSvc := ghClient.Git()

	// list_tags
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_tags",
		Description: "列出仓库的 Tags",
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
		perPage := request.GetInt("per_page", 30)

		tags, err := gitSvc.ListTags(ctx, owner, repo, &gogithub.ListOptions{PerPage: perPage})
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatTagListSimple(tags)), nil
	})

	// create_tag
	mcpServer.AddTool(mcp.Tool{
		Name:        "create_tag",
		Description: "创建 Tag（轻量 Tag 或附注 Tag）",
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
				"tag_name": map[string]any{
					"type":        "string",
					"description": "Tag 名称",
				},
				"target_commitish": map[string]any{
					"type":        "string",
					"description": "目标 commit SHA 或分支名",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "Tag 消息（创建附注 Tag）",
				},
			},
			Required: []string{"repo", "tag_name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		tagName := request.GetString("tag_name", "")
		target := request.GetString("target_commitish", "")
		message := request.GetString("message", "")

		req := &gogithub.RepositoryRelease{
			TagName: &tagName,
		}
		if target != "" {
			req.TargetCommitish = &target
		}
		if message != "" {
			req.Name = &message
		}

		// 使用 CreateRelease 来创建 tag（GitHub API 的 release 会自动创建 tag）
		release, err := releaseSvc.CreateRelease(ctx, owner, repo, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("✅ 已创建 Tag: %s\n\n%s", release.TagName, github.FormatRelease(release))), nil
	})

	// list_releases
	mcpServer.AddTool(mcp.Tool{
		Name:        "list_releases",
		Description: "列出仓库的 Releases",
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
		perPage := request.GetInt("per_page", 30)

		releases, err := releaseSvc.ListReleases(ctx, owner, repo, &gogithub.ListOptions{PerPage: perPage})
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatReleaseList(releases, owner, repo)), nil
	})

	// create_release
	mcpServer.AddTool(mcp.Tool{
		Name:        "create_release",
		Description: "创建 Release。可自动生成 Release Notes",
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
				"tag_name": map[string]any{
					"type":        "string",
					"description": "Tag 名称",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Release 标题",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Release 描述",
				},
				"draft": map[string]any{
					"type":        "boolean",
					"description": "是否为草稿，默认 false",
				},
				"prerelease": map[string]any{
					"type":        "boolean",
					"description": "是否为预发布，默认 false",
				},
				"generate_notes": map[string]any{
					"type":        "boolean",
					"description": "是否自动生成 Release Notes，默认 false",
				},
			},
			Required: []string{"repo", "tag_name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		tagName := request.GetString("tag_name", "")
		name := request.GetString("name", "")
		body := request.GetString("body", "")
		draft := request.GetBool("draft", false)
		prerelease := request.GetBool("prerelease", false)
		generateNotes := request.GetBool("generate_notes", false)

		req := &gogithub.RepositoryRelease{
			TagName: &tagName,
			Draft:   &draft,
			Prerelease: &prerelease,
		}
		if name != "" {
			req.Name = &name
		}
		if body != "" {
			req.Body = &body
		}

		release, err := releaseSvc.CreateRelease(ctx, owner, repo, req)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		result := fmt.Sprintf("✅ 已创建 Release: %s\n\n%s", release.TagName, github.FormatRelease(release))

		// 如果需要自动生成 Release Notes
		if generateNotes {
			notes, err := releaseSvc.GenerateReleaseNotes(ctx, owner, repo, tagName, "")
			if err == nil {
				result += "\n\n## 自动生成的 Release Notes\n\n" + notes
			}
		}

		return textResult(result), nil
	})

	// generate_release_notes
	mcpServer.AddTool(mcp.Tool{
		Name:        "generate_release_notes",
		Description: "基于 conventional commits 自动生成 Release Notes",
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
				"tag_name": map[string]any{
					"type":        "string",
					"description": "目标 Tag 名称",
				},
				"previous_tag": map[string]any{
					"type":        "string",
					"description": "对比基准 Tag（默认自动推断）",
				},
			},
			Required: []string{"repo", "tag_name"},
		},
	}, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		owner := request.GetString("owner", "")
		repo := request.GetString("repo", "")
		tagName := request.GetString("tag_name", "")
		previousTag := request.GetString("previous_tag", "")

		notes, err := releaseSvc.GenerateReleaseNotes(ctx, owner, repo, tagName, previousTag)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(fmt.Sprintf("## Release Notes for %s\n\n%s", tagName, notes)), nil
	})

	// get_latest_release
	mcpServer.AddTool(mcp.Tool{
		Name:        "get_latest_release",
		Description: "获取仓库最新的 Release",
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

		release, err := releaseSvc.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		return textResult(github.FormatRelease(release)), nil
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
