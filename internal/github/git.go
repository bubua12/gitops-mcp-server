package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
)

// GitService Git 操作服务
type GitService struct {
	client *Client
}

// CommitInfo Commit 信息
type CommitInfo struct {
	SHA     string
	Message string
	Author  string
	Date    string
	URL     string
}

// TagInfo Tag 信息
type TagInfo struct {
	Name   string
	SHA    string
	Commit string
}

// ListCommits 列出 commits
func (s *GitService) ListCommits(ctx context.Context, owner, repo string, opts *github.CommitsListOptions) ([]*CommitInfo, error) {
	owner = s.client.resolveOwner(owner)
	commits, _, err := s.client.gh.Repositories.ListCommits(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list commits %s/%s: %w", owner, repo, err)
	}
	var result []*CommitInfo
	for _, c := range commits {
		result = append(result, &CommitInfo{
			SHA:     c.GetSHA(),
			Message: c.GetCommit().GetMessage(),
			Author:  c.GetCommit().GetAuthor().GetName(),
			Date:    c.GetCommit().GetAuthor().GetDate().String(),
			URL:     c.GetHTMLURL(),
		})
	}
	return result, nil
}

// ListTags 列出 tags
func (s *GitService) ListTags(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*TagInfo, error) {
	owner = s.client.resolveOwner(owner)
	tags, _, err := s.client.gh.Repositories.ListTags(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list tags %s/%s: %w", owner, repo, err)
	}
	var result []*TagInfo
	for _, t := range tags {
		result = append(result, &TagInfo{
			Name:   t.GetName(),
			SHA:    t.GetCommit().GetSHA(),
			Commit: t.GetCommit().GetSHA(),
		})
	}
	return result, nil
}

// CompareRefs 对比两个 ref
func (s *GitService) CompareRefs(ctx context.Context, owner, repo, base, head string) (string, error) {
	owner = s.client.resolveOwner(owner)
	comparison, _, err := s.client.gh.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	if err != nil {
		return "", fmt.Errorf("compare refs %s...%s: %w", base, head, err)
	}

	return formatComparison(comparison, base, head), nil
}

func formatComparison(comp *github.CommitsComparison, base, head string) string {
	var result string
	result = fmt.Sprintf("## 对比: %s → %s\n\n", base, head)
	result += fmt.Sprintf("- **状态:** %s\n", comp.GetStatus())
	result += fmt.Sprintf("- **总提交数:** %d\n", comp.GetTotalCommits())
	result += fmt.Sprintf("- **领先:** %d | **落后:** %d\n", comp.GetAheadBy(), comp.GetBehindBy())
	result += fmt.Sprintf("- **文件变更数:** %d\n", len(comp.Files))

	// 计算总新增/删除行数
	totalAdditions, totalDeletions := 0, 0
	for _, f := range comp.Files {
		totalAdditions += f.GetAdditions()
		totalDeletions += f.GetDeletions()
	}
	result += fmt.Sprintf("- **新增行:** %d | **删除行:** %d\n", totalAdditions, totalDeletions)

	if len(comp.Commits) > 0 {
		result += "\n### 提交记录\n\n"
		for _, c := range comp.Commits {
			msg := c.GetCommit().GetMessage()
			if len(msg) > 80 {
				msg = msg[:80] + "..."
			}
			result += fmt.Sprintf("- `%s` %s (%s)\n",
				c.GetSHA()[:7], msg, c.GetCommit().GetAuthor().GetName())
		}
	}

	if len(comp.Files) > 0 {
		result += "\n### 变更文件\n\n"
		for _, f := range comp.Files {
			result += fmt.Sprintf("- `%s` (+%d/-%d) %s\n",
				f.GetFilename(), f.GetAdditions(), f.GetDeletions(), f.GetStatus())
		}
	}

	return result
}

// FormatCommitList 格式化 commit 列表
func FormatCommitList(commits []*CommitInfo) string {
	result := "📝 提交记录\n\n"
	if len(commits) == 0 {
		return result + "无提交记录\n"
	}
	for _, c := range commits {
		msg := c.Message
		if idx := indexOfNewline(msg); idx > 0 {
			msg = msg[:idx]
		}
		if len(msg) > 80 {
			msg = msg[:80] + "..."
		}
		result += fmt.Sprintf("- `%s` %s (%s, %s)\n", c.SHA[:7], msg, c.Author, c.Date)
	}
	return result
}

// FormatTagList 格式化 tag 列表
func FormatTagList(tags []*TagInfo) string {
	result := "🏷️ Tags\n\n"
	if len(tags) == 0 {
		return result + "无 Tag\n"
	}
	for _, t := range tags {
		result += fmt.Sprintf("- **%s** (`%s`)\n", t.Name, t.SHA[:7])
	}
	return result
}

func indexOfNewline(s string) int {
	for i, c := range s {
		if c == '\n' {
			return i
		}
	}
	return -1
}
