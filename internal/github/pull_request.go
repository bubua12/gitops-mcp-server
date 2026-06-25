package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

// PullRequestService PR 操作服务
type PullRequestService struct {
	client *Client
}

// PRInfo PR 信息
type PRInfo struct {
	Number       int
	Title        string
	State        string
	Body         string
	User         string
	Head         string
	Base         string
	URL          string
	Mergeable    bool
	Merged       bool
	Draft        bool
	Additions    int
	Deletions    int
	ChangedFiles int
	CreatedAt    string
	UpdatedAt    string
	Comments     int
	Reviewers    []string
}

// ReviewInfo 审查信息
type ReviewInfo struct {
	ID       int
	User     string
	State    string
	Body     string
	CommitID string
}

// DiffFile 变更文件信息
type DiffFile struct {
	Filename  string
	Status    string
	Additions int
	Deletions int
	Changes   int
	Patch     string
}

// List 列出 PR
func (s *PullRequestService) List(ctx context.Context, owner, repo string, opts *github.PullRequestListOptions) ([]*PRInfo, error) {
	owner = s.client.resolveOwner(owner)
	prs, _, err := s.client.gh.PullRequests.List(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list PRs %s/%s: %w", owner, repo, err)
	}
	var result []*PRInfo
	for _, pr := range prs {
		result = append(result, prToInfo(pr))
	}
	return result, nil
}

// Get 获取 PR 详情
func (s *PullRequestService) Get(ctx context.Context, owner, repo string, number int) (*PRInfo, error) {
	owner = s.client.resolveOwner(owner)
	pr, _, err := s.client.gh.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("get PR %s/%s#%d: %w", owner, repo, number, err)
	}
	return prToInfo(pr), nil
}

// GetDiff 获取 PR diff
func (s *PullRequestService) GetDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	owner = s.client.resolveOwner(owner)
	opt := github.RawOptions{Type: github.Diff}
	raw, _, err := s.client.gh.PullRequests.GetRaw(ctx, owner, repo, number, opt)
	if err != nil {
		return "", fmt.Errorf("get PR diff %s/%s#%d: %w", owner, repo, number, err)
	}
	return string(raw), nil
}

// GetFiles 获取 PR 变更文件列表
func (s *PullRequestService) GetFiles(ctx context.Context, owner, repo string, number int) ([]*DiffFile, error) {
	owner = s.client.resolveOwner(owner)
	files, _, err := s.client.gh.PullRequests.ListFiles(ctx, owner, repo, number, nil)
	if err != nil {
		return nil, fmt.Errorf("get PR files %s/%s#%d: %w", owner, repo, number, err)
	}
	var result []*DiffFile
	for _, f := range files {
		result = append(result, &DiffFile{
			Filename:  f.GetFilename(),
			Status:    f.GetStatus(),
			Additions: f.GetAdditions(),
			Deletions: f.GetDeletions(),
			Changes:   f.GetChanges(),
			Patch:     f.GetPatch(),
		})
	}
	return result, nil
}

// GetReviews 获取 PR 审查历史
func (s *PullRequestService) GetReviews(ctx context.Context, owner, repo string, number int) ([]*ReviewInfo, error) {
	owner = s.client.resolveOwner(owner)
	reviews, _, err := s.client.gh.PullRequests.ListReviews(ctx, owner, repo, number, nil)
	if err != nil {
		return nil, fmt.Errorf("get PR reviews %s/%s#%d: %w", owner, repo, number, err)
	}
	var result []*ReviewInfo
	for _, r := range reviews {
		result = append(result, &ReviewInfo{
			ID:       int(r.GetID()),
			User:     r.GetUser().GetLogin(),
			State:    r.GetState(),
			Body:     r.GetBody(),
			CommitID: r.GetCommitID(),
		})
	}
	return result, nil
}

// Create 创建 PR
func (s *PullRequestService) Create(ctx context.Context, owner, repo string, req *github.NewPullRequest) (*PRInfo, error) {
	owner = s.client.resolveOwner(owner)
	pr, _, err := s.client.gh.PullRequests.Create(ctx, owner, repo, req)
	if err != nil {
		return nil, fmt.Errorf("create PR %s/%s: %w", owner, repo, err)
	}
	return prToInfo(pr), nil
}

// CreateReview 提交代码审查
func (s *PullRequestService) CreateReview(ctx context.Context, owner, repo string, number int, req *github.PullRequestReviewRequest) (*ReviewInfo, error) {
	owner = s.client.resolveOwner(owner)
	review, _, err := s.client.gh.PullRequests.CreateReview(ctx, owner, repo, number, req)
	if err != nil {
		return nil, fmt.Errorf("create review for %s/%s#%d: %w", owner, repo, number, err)
	}
	return &ReviewInfo{
		ID:       int(review.GetID()),
		User:     review.GetUser().GetLogin(),
		State:    review.GetState(),
		Body:     review.GetBody(),
		CommitID: review.GetCommitID(),
	}, nil
}

// Merge 合并 PR
func (s *PullRequestService) Merge(ctx context.Context, owner, repo string, number int, req *github.PullRequestOptions) (string, error) {
	owner = s.client.resolveOwner(owner)
	result, _, err := s.client.gh.PullRequests.Merge(ctx, owner, repo, number, "", req)
	if err != nil {
		return "", fmt.Errorf("merge PR %s/%s#%d: %w", owner, repo, number, err)
	}
	return result.GetMessage(), nil
}

// RequestReviewers 请求审查者
func (s *PullRequestService) RequestReviewers(ctx context.Context, owner, repo string, number int, reviewers []string) error {
	owner = s.client.resolveOwner(owner)
	req := github.ReviewersRequest{
		Reviewers: reviewers,
	}
	_, _, err := s.client.gh.PullRequests.RequestReviewers(ctx, owner, repo, number, req)
	if err != nil {
		return fmt.Errorf("request reviewers for %s/%s#%d: %w", owner, repo, number, err)
	}
	return nil
}

func prToInfo(pr *github.PullRequest) *PRInfo {
	reviewers := make([]string, len(pr.RequestedReviewers))
	for i, r := range pr.RequestedReviewers {
		reviewers[i] = r.GetLogin()
	}
	return &PRInfo{
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		State:        pr.GetState(),
		Body:         pr.GetBody(),
		User:         pr.GetUser().GetLogin(),
		Head:         pr.GetHead().GetRef(),
		Base:         pr.GetBase().GetRef(),
		URL:          pr.GetHTMLURL(),
		Mergeable:    pr.GetMergeable(),
		Merged:       pr.GetMerged(),
		Draft:        pr.GetDraft(),
		Additions:    pr.GetAdditions(),
		Deletions:    pr.GetDeletions(),
		ChangedFiles: pr.GetChangedFiles(),
		CreatedAt:    pr.GetCreatedAt().String(),
		UpdatedAt:    pr.GetUpdatedAt().String(),
		Comments:     pr.GetComments(),
		Reviewers:    reviewers,
	}
}

// FormatPR 格式化 PR 为 Markdown
func FormatPR(pr *PRInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# PR #%d %s\n\n", pr.Number, pr.Title))

	state := "🟢 Open"
	if pr.State == "closed" {
		if pr.Merged {
			state = "🟣 Merged"
		} else {
			state = "🔴 Closed"
		}
	}
	if pr.Draft {
		state += " (Draft)"
	}

	sb.WriteString(fmt.Sprintf("- **状态:** %s\n", state))
	sb.WriteString(fmt.Sprintf("- **作者:** %s\n", pr.User))
	sb.WriteString(fmt.Sprintf("- **分支:** `%s` → `%s`\n", pr.Head, pr.Base))
	sb.WriteString(fmt.Sprintf("- **变更:** +%d/-%d (%d files)\n", pr.Additions, pr.Deletions, pr.ChangedFiles))
	sb.WriteString(fmt.Sprintf("- **链接:** %s\n", pr.URL))

	if len(pr.Reviewers) > 0 {
		sb.WriteString(fmt.Sprintf("- **审查者:** %s\n", strings.Join(pr.Reviewers, ", ")))
	}

	if pr.Body != "" {
		sb.WriteString(fmt.Sprintf("\n## 描述\n\n%s\n", pr.Body))
	}
	return sb.String()
}

// FormatPRList 格式化 PR 列表
func FormatPRList(prs []*PRInfo, owner, repo string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔀 Pull Requests: %s/%s\n\n", owner, repo))
	if len(prs) == 0 {
		sb.WriteString("无 Pull Requests\n")
		return sb.String()
	}
	for _, pr := range prs {
		state := "🟢"
		if pr.State == "closed" {
			if pr.Merged {
				state = "🟣"
			} else {
				state = "🔴"
			}
		}
		draft := ""
		if pr.Draft {
			draft = " [Draft]"
		}
		sb.WriteString(fmt.Sprintf("%s #%d %s (%s → %s)%s\n", state, pr.Number, pr.Title, pr.Head, pr.Base, draft))
	}
	return sb.String()
}

// FormatDiffFiles 格式化变更文件列表
func FormatDiffFiles(files []*DiffFile) string {
	var sb strings.Builder
	sb.WriteString("📁 变更文件列表\n\n")
	if len(files) == 0 {
		sb.WriteString("无变更\n")
		return sb.String()
	}
	for _, f := range files {
		sb.WriteString(fmt.Sprintf("- `%s` (+%d/-%d) %s\n", f.Filename, f.Additions, f.Deletions, f.Status))
	}
	return sb.String()
}

// FormatReviewList 格式化审查列表
func FormatReviewList(reviews []*ReviewInfo) string {
	var sb strings.Builder
	sb.WriteString("👀 审查历史\n\n")
	if len(reviews) == 0 {
		sb.WriteString("无审查记录\n")
		return sb.String()
	}
	for _, r := range reviews {
		sb.WriteString(fmt.Sprintf("- **%s** [%s]: %s\n", r.User, r.State, r.Body))
	}
	return sb.String()
}
