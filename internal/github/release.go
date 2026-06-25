package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

// ReleaseService Release 操作服务
type ReleaseService struct {
	client *Client
}

// ReleaseInfo Release 信息
type ReleaseInfo struct {
	ID          int
	TagName     string
	Name        string
	Body        string
	Draft       bool
	Prerelease  bool
	URL         string
	HTMLURL     string
	TarballURL  string
	ZipballURL  string
	PublishedAt string
	CreatedAt   string
	Author      string
}

// ListReleases 列出 Releases
func (s *ReleaseService) ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*ReleaseInfo, error) {
	owner = s.client.resolveOwner(owner)
	releases, _, err := s.client.gh.Repositories.ListReleases(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list releases %s/%s: %w", owner, repo, err)
	}
	var result []*ReleaseInfo
	for _, r := range releases {
		result = append(result, releaseToInfo(r))
	}
	return result, nil
}

// GetRelease 获取 Release 详情
func (s *ReleaseService) GetRelease(ctx context.Context, owner, repo string, id int64) (*ReleaseInfo, error) {
	owner = s.client.resolveOwner(owner)
	release, _, err := s.client.gh.Repositories.GetRelease(ctx, owner, repo, id)
	if err != nil {
		return nil, fmt.Errorf("get release %s/%s: %w", owner, repo, err)
	}
	return releaseToInfo(release), nil
}

// GetLatestRelease 获取最新 Release
func (s *ReleaseService) GetLatestRelease(ctx context.Context, owner, repo string) (*ReleaseInfo, error) {
	owner = s.client.resolveOwner(owner)
	release, _, err := s.client.gh.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("get latest release %s/%s: %w", owner, repo, err)
	}
	return releaseToInfo(release), nil
}

// CreateRelease 创建 Release
func (s *ReleaseService) CreateRelease(ctx context.Context, owner, repo string, req *github.RepositoryRelease) (*ReleaseInfo, error) {
	owner = s.client.resolveOwner(owner)
	release, _, err := s.client.gh.Repositories.CreateRelease(ctx, owner, repo, req)
	if err != nil {
		return nil, fmt.Errorf("create release %s/%s: %w", owner, repo, err)
	}
	return releaseToInfo(release), nil
}

// UpdateRelease 更新 Release
func (s *ReleaseService) UpdateRelease(ctx context.Context, owner, repo string, id int64, req *github.RepositoryRelease) (*ReleaseInfo, error) {
	owner = s.client.resolveOwner(owner)
	release, _, err := s.client.gh.Repositories.EditRelease(ctx, owner, repo, id, req)
	if err != nil {
		return nil, fmt.Errorf("update release %s/%s: %w", owner, repo, err)
	}
	return releaseToInfo(release), nil
}

// DeleteRelease 删除 Release
func (s *ReleaseService) DeleteRelease(ctx context.Context, owner, repo string, id int64) error {
	owner = s.client.resolveOwner(owner)
	_, err := s.client.gh.Repositories.DeleteRelease(ctx, owner, repo, id)
	if err != nil {
		return fmt.Errorf("delete release %s/%s: %w", owner, repo, err)
	}
	return nil
}

// GenerateReleaseNotes 生成 Release Notes
func (s *ReleaseService) GenerateReleaseNotes(ctx context.Context, owner, repo, tagName, previousTag string) (string, error) {
	owner = s.client.resolveOwner(owner)
	opts := &github.GenerateNotesOptions{
		TagName: tagName,
	}
	if previousTag != "" {
		opts.PreviousTagName = &previousTag
	}
	notes, _, err := s.client.gh.Repositories.GenerateReleaseNotes(ctx, owner, repo, opts)
	if err != nil {
		return "", fmt.Errorf("generate release notes %s/%s: %w", owner, repo, err)
	}
	return notes.Body, nil
}

func releaseToInfo(r *github.RepositoryRelease) *ReleaseInfo {
	return &ReleaseInfo{
		ID:          int(r.GetID()),
		TagName:     r.GetTagName(),
		Name:        r.GetName(),
		Body:        r.GetBody(),
		Draft:       r.GetDraft(),
		Prerelease:  r.GetPrerelease(),
		URL:         r.GetURL(),
		HTMLURL:     r.GetHTMLURL(),
		TarballURL:  r.GetTarballURL(),
		ZipballURL:  r.GetZipballURL(),
		PublishedAt: r.GetPublishedAt().String(),
		CreatedAt:   r.GetCreatedAt().String(),
		Author:      r.GetAuthor().GetLogin(),
	}
}

// FormatRelease 格式化 Release 为 Markdown
func FormatRelease(r *ReleaseInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", r.Name))

	status := ""
	if r.Draft {
		status = "📝 Draft"
	} else if r.Prerelease {
		status = "🧪 Pre-release"
	} else {
		status = "✅ Release"
	}

	sb.WriteString(fmt.Sprintf("- **Tag:** %s\n", r.TagName))
	sb.WriteString(fmt.Sprintf("- **状态:** %s\n", status))
	sb.WriteString(fmt.Sprintf("- **作者:** %s\n", r.Author))
	sb.WriteString(fmt.Sprintf("- **发布时间:** %s\n", r.PublishedAt))
	sb.WriteString(fmt.Sprintf("- **链接:** %s\n", r.HTMLURL))

	if r.Body != "" {
		sb.WriteString(fmt.Sprintf("\n## Release Notes\n\n%s\n", r.Body))
	}
	return sb.String()
}

// FormatReleaseList 格式化 Release 列表
func FormatReleaseList(releases []*ReleaseInfo, owner, repo string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚀 Releases: %s/%s\n\n", owner, repo))
	if len(releases) == 0 {
		sb.WriteString("无 Release\n")
		return sb.String()
	}
	for _, r := range releases {
		status := "✅"
		if r.Draft {
			status = "📝"
		} else if r.Prerelease {
			status = "🧪"
		}
		sb.WriteString(fmt.Sprintf("%s **%s** (%s) - %s\n", status, r.Name, r.TagName, r.PublishedAt))
	}
	return sb.String()
}

// FormatTagList 格式化 Tag 列表
func FormatTagListSimple(tags []*TagInfo) string {
	var sb strings.Builder
	sb.WriteString("🏷️ Tags\n\n")
	if len(tags) == 0 {
		sb.WriteString("无 Tag\n")
		return sb.String()
	}
	for _, t := range tags {
		sb.WriteString(fmt.Sprintf("- **%s** (`%s`)\n", t.Name, t.SHA[:7]))
	}
	return sb.String()
}

// CommitToReleaseNotes 将 commits 转换为 Release Notes
func CommitToReleaseNotes(commits []*CommitInfo, tagName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## %s\n\n", tagName))

	// 按 conventional commits 分类
	var features, fixes, docs, others []string
	for _, c := range commits {
		msg := c.Message
		if idx := indexOfNewline(msg); idx > 0 {
			msg = msg[:idx]
		}
		msg = strings.TrimSpace(msg)

		switch {
		case strings.HasPrefix(msg, "feat"):
			features = append(features, msg)
		case strings.HasPrefix(msg, "fix"):
			fixes = append(fixes, msg)
		case strings.HasPrefix(msg, "docs"):
			docs = append(docs, msg)
		default:
			others = append(others, msg)
		}
	}

	if len(features) > 0 {
		sb.WriteString("### ✨ 新功能\n\n")
		for _, f := range features {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(fixes) > 0 {
		sb.WriteString("### 🐛 Bug 修复\n\n")
		for _, f := range fixes {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(docs) > 0 {
		sb.WriteString("### 📚 文档\n\n")
		for _, d := range docs {
			sb.WriteString(fmt.Sprintf("- %s\n", d))
		}
		sb.WriteString("\n")
	}

	if len(others) > 0 {
		sb.WriteString("### 🔧 其他变更\n\n")
		for _, o := range others {
			sb.WriteString(fmt.Sprintf("- %s\n", o))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
