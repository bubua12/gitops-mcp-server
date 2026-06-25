package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v68/github"
)

// RepositoryService 仓库操作服务
type RepositoryService struct {
	client *Client
}

// RepoInfo 仓库信息
type RepoInfo struct {
	Name        string
	FullName    string
	Description string
	Language    string
	Stars       int
	Forks       int
	OpenIssues  int
	DefaultBranch string
	URL         string
	CreatedAt   string
	UpdatedAt   string
	Private     bool
	Archived    bool
}

// Get 获取仓库详情
func (s *RepositoryService) Get(ctx context.Context, owner, repo string) (*RepoInfo, error) {
	owner = s.client.resolveOwner(owner)
	r, _, err := s.client.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get repo %s/%s: %w", owner, repo, err)
	}
	return repoToInfo(r), nil
}

// List 列出用户的仓库
func (s *RepositoryService) List(ctx context.Context, owner string, opts *github.RepositoryListOptions) ([]*RepoInfo, error) {
	owner = s.client.resolveOwner(owner)
	repos, _, err := s.client.gh.Repositories.List(ctx, owner, opts)
	if err != nil {
		return nil, fmt.Errorf("list repos for %s: %w", owner, err)
	}
	var result []*RepoInfo
	for _, r := range repos {
		result = append(result, repoToInfo(r))
	}
	return result, nil
}

// Search 搜索仓库
func (s *RepositoryService) Search(ctx context.Context, query string, opts *github.SearchOptions) ([]*RepoInfo, int, error) {
	result, _, err := s.client.gh.Search.Repositories(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("search repos: %w", err)
	}
	var repos []*RepoInfo
	for _, r := range result.Repositories {
		repos = append(repos, repoToInfo(r))
	}
	return repos, result.GetTotal(), nil
}

// GetContent 获取仓库文件/目录内容
func (s *RepositoryService) GetContent(ctx context.Context, owner, repo, path, ref string) (string, string, []*github.RepositoryContent, error) {
	owner = s.client.resolveOwner(owner)

	fileContent, dirContent, resp, err := s.client.gh.Repositories.GetContents(ctx, owner, repo, path, &github.RepositoryContentGetOptions{
		Ref: ref,
	})
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", "", nil, fmt.Errorf("path not found: %s", path)
		}
		return "", "", nil, fmt.Errorf("get content %s/%s/%s: %w", owner, repo, path, err)
	}

	if fileContent != nil {
		content, err := fileContent.GetContent()
		if err != nil {
			return "", "", nil, fmt.Errorf("decode content: %w", err)
		}
		return content, fileContent.GetSHA(), nil, nil
	}

	return "", "", dirContent, nil
}

// SearchCode 搜索代码
func (s *RepositoryService) SearchCode(ctx context.Context, query string, owner, repo string, opts *github.SearchOptions) ([]CodeSearchResult, int, error) {
	q := query
	if owner != "" && repo != "" {
		q = fmt.Sprintf("%s repo:%s/%s", query, s.client.resolveOwner(owner), repo)
	} else if owner != "" {
		q = fmt.Sprintf("%s user:%s", query, s.client.resolveOwner(owner))
	}

	result, _, err := s.client.gh.Search.Code(ctx, q, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("search code: %w", err)
	}
	var results []CodeSearchResult
	for _, r := range result.CodeResults {
		results = append(results, CodeSearchResult{
			Name:     r.GetName(),
			Path:     r.GetPath(),
			URL:      r.GetHTMLURL(),
			RepoName: r.Repository.GetFullName(),
		})
	}
	return results, result.GetTotal(), nil
}

// GetTree 获取仓库目录树
func (s *RepositoryService) GetTree(ctx context.Context, owner, repo, sha string, recursive bool) ([]TreeEntry, error) {
	owner = s.client.resolveOwner(owner)
	tree, _, err := s.client.gh.Git.GetTree(ctx, owner, repo, sha, recursive)
	if err != nil {
		return nil, fmt.Errorf("get tree: %w", err)
	}
	var entries []TreeEntry
	for _, e := range tree.Entries {
		entries = append(entries, TreeEntry{
			Path: e.GetPath(),
			Type: e.GetType(),
			SHA:  e.GetSHA(),
			Size: e.GetSize(),
			URL:  e.GetURL(),
		})
	}
	return entries, nil
}

// CodeSearchResult 代码搜索结果
type CodeSearchResult struct {
	Name     string
	Path     string
	URL      string
	RepoName string
}

// TreeEntry 目录树条目
type TreeEntry struct {
	Path string
	Type string
	SHA  string
	Size int
	URL  string
}

func repoToInfo(r *github.Repository) *RepoInfo {
	return &RepoInfo{
		Name:          r.GetName(),
		FullName:      r.GetFullName(),
		Description:   r.GetDescription(),
		Language:      r.GetLanguage(),
		Stars:         r.GetStargazersCount(),
		Forks:         r.GetForksCount(),
		OpenIssues:    r.GetOpenIssuesCount(),
		DefaultBranch: r.GetDefaultBranch(),
		URL:           r.GetHTMLURL(),
		CreatedAt:     r.GetCreatedAt().String(),
		UpdatedAt:     r.GetUpdatedAt().String(),
		Private:       r.GetPrivate(),
		Archived:      r.GetArchived(),
	}
}

// FormatRepoInfo 格式化仓库信息为 Markdown
func FormatRepoInfo(info *RepoInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", info.FullName))
	if info.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", info.Description))
	}
	sb.WriteString(fmt.Sprintf("- **语言:** %s\n", info.Language))
	sb.WriteString(fmt.Sprintf("- **Stars:** %d | **Forks:** %d | **Open Issues:** %d\n", info.Stars, info.Forks, info.OpenIssues))
	sb.WriteString(fmt.Sprintf("- **默认分支:** %s\n", info.DefaultBranch))
	sb.WriteString(fmt.Sprintf("- **URL:** %s\n", info.URL))
	if info.Archived {
		sb.WriteString("- **状态:** ⚠️ 已归档\n")
	}
	return sb.String()
}

// FormatTree 格式化目录树为可读文本
func FormatTree(entries []TreeEntry, rootPath string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📁 目录结构: %s\n\n", rootPath))
	for _, e := range entries {
		indent := ""
		depth := strings.Count(e.Path, "/")
		if depth > 0 {
			indent = strings.Repeat("  ", depth)
		}
		if e.Type == "tree" {
			sb.WriteString(fmt.Sprintf("%s📁 %s/\n", indent, e.Path[strings.LastIndex(e.Path, "/")+1:]))
		} else {
			sb.WriteString(fmt.Sprintf("%s📄 %s\n", indent, e.Path[strings.LastIndex(e.Path, "/")+1:]))
		}
	}
	return sb.String()
}
