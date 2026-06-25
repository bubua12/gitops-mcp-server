package github

import (
	"context"
	"fmt"
	"strings"

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

// CommitDetail Commit 详细信息
type CommitDetail struct {
	CommitInfo
	Stats      *CommitStats
	Files      []*CommitFileInfo
	Parents    []string
}

// CommitStats Commit 统计
type CommitStats struct {
	Additions int
	Deletions int
	Total     int
}

// CommitFileInfo Commit 文件变更信息
type CommitFileInfo struct {
	Filename  string
	Status    string
	Additions int
	Deletions int
	Changes   int
	Patch     string
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

// GetCommit 获取 commit 详情
func (s *GitService) GetCommit(ctx context.Context, owner, repo, sha string) (*CommitDetail, error) {
	owner = s.client.resolveOwner(owner)
	commit, _, err := s.client.gh.Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("get commit %s/%s@%s: %w", owner, repo, sha, err)
	}

	detail := &CommitDetail{
		CommitInfo: CommitInfo{
			SHA:     commit.GetSHA(),
			Message: commit.GetCommit().GetMessage(),
			Author:  commit.GetCommit().GetAuthor().GetName(),
			Date:    commit.GetCommit().GetAuthor().GetDate().String(),
			URL:     commit.GetHTMLURL(),
		},
		Stats: &CommitStats{
			Additions: commit.GetStats().GetAdditions(),
			Deletions: commit.GetStats().GetDeletions(),
			Total:     commit.GetStats().GetTotal(),
		},
	}

	// Parents
	for _, p := range commit.Parents {
		detail.Parents = append(detail.Parents, p.GetSHA())
	}

	// Files
	for _, f := range commit.Files {
		detail.Files = append(detail.Files, &CommitFileInfo{
			Filename:  f.GetFilename(),
			Status:    f.GetStatus(),
			Additions: f.GetAdditions(),
			Deletions: f.GetDeletions(),
			Changes:   f.GetChanges(),
			Patch:     f.GetPatch(),
		})
	}

	return detail, nil
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

// CreateLightweightTag 创建轻量 Tag（指向 commit）
func (s *GitService) CreateLightweightTag(ctx context.Context, owner, repo, tagName, sha string) error {
	owner = s.client.resolveOwner(owner)
	ref := "refs/tags/" + tagName
	_, _, err := s.client.gh.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    &ref,
		Object: &github.GitObject{SHA: &sha},
	})
	if err != nil {
		return fmt.Errorf("create lightweight tag %s: %w", tagName, err)
	}
	return nil
}

// CreateAnnotatedTag 创建附注 Tag（带消息）
func (s *GitService) CreateAnnotatedTag(ctx context.Context, owner, repo, tagName, message, sha string) error {
	owner = s.client.resolveOwner(owner)
	tagType := "commit"
	tag, _, err := s.client.gh.Git.CreateTag(ctx, owner, repo, &github.Tag{
		Tag:     &tagName,
		Message: &message,
		Object: &github.GitObject{
			SHA:  &sha,
			Type: &tagType,
		},
	})
	if err != nil {
		return fmt.Errorf("create annotated tag %s: %w", tagName, err)
	}
	// 创建 ref 指向附注 Tag 对象
	ref := "refs/tags/" + tagName
	_, _, err = s.client.gh.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    &ref,
		Object: &github.GitObject{SHA: tag.SHA},
	})
	if err != nil {
		return fmt.Errorf("create ref for tag %s: %w", tagName, err)
	}
	return nil
}

// GetDefaultBranchSHA 获取默认分支最新 commit SHA
func (s *GitService) GetDefaultBranchSHA(ctx context.Context, owner, repo string) (string, error) {
	owner = s.client.resolveOwner(owner)
	r, _, err := s.client.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("get repo %s/%s: %w", owner, repo, err)
	}
	branch := r.GetDefaultBranch()
	ref, _, err := s.client.gh.Git.GetRef(ctx, owner, repo, "refs/heads/"+branch)
	if err != nil {
		return "", fmt.Errorf("get default branch ref: %w", err)
	}
	return ref.GetObject().GetSHA(), nil
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

// AnalyzeChanges 分析变更影响
func (s *GitService) AnalyzeChanges(ctx context.Context, owner, repo, base, head string) (string, error) {
	owner = s.client.resolveOwner(owner)
	comp, _, err := s.client.gh.Repositories.CompareCommits(ctx, owner, repo, base, head, nil)
	if err != nil {
		return "", fmt.Errorf("analyze changes %s...%s: %w", base, head, err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("## 变更分析: %s → %s\n\n", base, head))

	// 基本统计
	totalAdditions, totalDeletions := 0, 0
	fileTypes := make(map[string]int)
	affectedDirs := make(map[string]bool)

	for _, f := range comp.Files {
		totalAdditions += f.GetAdditions()
		totalDeletions += f.GetDeletions()

		// 文件类型统计
		if idx := strings.LastIndex(f.GetFilename(), "."); idx >= 0 {
			ext := f.GetFilename()[idx:]
			fileTypes[ext]++
		}

		// 影响目录
		if idx := strings.LastIndex(f.GetFilename(), "/"); idx >= 0 {
			dir := f.GetFilename()[:idx]
			affectedDirs[dir] = true
		}
	}

	result.WriteString("### 📊 统计概览\n\n")
	result.WriteString(fmt.Sprintf("- 提交数: %d\n", comp.GetTotalCommits()))
	result.WriteString(fmt.Sprintf("- 文件变更: %d\n", len(comp.Files)))
	result.WriteString(fmt.Sprintf("- 新增: +%d 行\n", totalAdditions))
	result.WriteString(fmt.Sprintf("- 删除: -%d 行\n", totalDeletions))
	result.WriteString(fmt.Sprintf("- 净变更: %d 行\n", totalAdditions-totalDeletions))

	// 文件类型分布
	if len(fileTypes) > 0 {
		result.WriteString("\n### 📁 文件类型分布\n\n")
		for ext, count := range fileTypes {
			result.WriteString(fmt.Sprintf("- %s: %d 个文件\n", ext, count))
		}
	}

	// 影响目录
	if len(affectedDirs) > 0 {
		result.WriteString("\n### 📂 影响目录\n\n")
		for dir := range affectedDirs {
			result.WriteString(fmt.Sprintf("- %s/\n", dir))
		}
	}

	// 大文件警告
	result.WriteString("\n### ⚠️ 风险提示\n\n")
	hasRisk := false
	for _, f := range comp.Files {
		total := f.GetAdditions() + f.GetDeletions()
		if total > 500 {
			result.WriteString(fmt.Sprintf("- **大变更**: %s (%d 行变更)\n", f.GetFilename(), total))
			hasRisk = true
		}
	}
	if !hasRisk {
		result.WriteString("- 无明显风险\n")
	}

	return result.String(), nil
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

// FormatCommitDetail 格式化 commit 详情
func FormatCommitDetail(detail *CommitDetail) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Commit %s\n\n", detail.SHA[:7]))

	// 基本信息
	sb.WriteString(fmt.Sprintf("- **作者:** %s\n", detail.Author))
	sb.WriteString(fmt.Sprintf("- **时间:** %s\n", detail.Date))
	sb.WriteString(fmt.Sprintf("- **链接:** %s\n", detail.URL))

	if detail.Stats != nil {
		sb.WriteString(fmt.Sprintf("- **变更:** +%d/-%d (共 %d 行)\n",
			detail.Stats.Additions, detail.Stats.Deletions, detail.Stats.Total))
	}

	if len(detail.Parents) > 0 {
		parents := make([]string, len(detail.Parents))
		for i, p := range detail.Parents {
			parents[i] = p[:7]
		}
		sb.WriteString(fmt.Sprintf("- **父提交:** %s\n", strings.Join(parents, ", ")))
	}

	// 提交信息
	sb.WriteString(fmt.Sprintf("\n### 提交信息\n\n%s\n", detail.Message))

	// 变更文件
	if len(detail.Files) > 0 {
		sb.WriteString("\n### 变更文件\n\n")
		for _, f := range detail.Files {
			sb.WriteString(fmt.Sprintf("- `%s` (+%d/-%d) %s\n",
				f.Filename, f.Additions, f.Deletions, f.Status))
		}
	}

	return sb.String()
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
