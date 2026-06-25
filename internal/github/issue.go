package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v68/github"
)

// IssueService Issue 操作服务
type IssueService struct {
	client *Client
}

// IssueInfo Issue 信息
type IssueInfo struct {
	Number    int
	Title     string
	State     string
	Body      string
	User      string
	Labels    []string
	Assignees []string
	URL       string
	CreatedAt string
	UpdatedAt string
	Comments  int
}

// CommentInfo 评论信息
type CommentInfo struct {
	ID        int
	Body      string
	User      string
	CreatedAt string
	UpdatedAt string
}

// List 列出 Issues
func (s *IssueService) List(ctx context.Context, owner, repo string, opts *github.IssueListByRepoOptions) ([]*IssueInfo, error) {
	owner = s.client.resolveOwner(owner)
	issues, _, err := s.client.gh.Issues.ListByRepo(ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("list issues %s/%s: %w", owner, repo, err)
	}
	var result []*IssueInfo
	for _, issue := range issues {
		if issue.IsPullRequest() {
			continue // 跳过 PR
		}
		result = append(result, issueToInfo(issue))
	}
	return result, nil
}

// Get 获取 Issue 详情
func (s *IssueService) Get(ctx context.Context, owner, repo string, number int) (*IssueInfo, error) {
	owner = s.client.resolveOwner(owner)
	issue, _, err := s.client.gh.Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("get issue %s/%s#%d: %w", owner, repo, number, err)
	}
	return issueToInfo(issue), nil
}

// ListComments 列出 Issue 评论
func (s *IssueService) ListComments(ctx context.Context, owner, repo string, number int, since *time.Time) ([]*CommentInfo, error) {
	owner = s.client.resolveOwner(owner)
	opts := &github.IssueListCommentsOptions{
		Since: since,
	}
	comments, _, err := s.client.gh.Issues.ListComments(ctx, owner, repo, number, opts)
	if err != nil {
		return nil, fmt.Errorf("list comments for %s/%s#%d: %w", owner, repo, number, err)
	}
	var result []*CommentInfo
	for _, c := range comments {
		result = append(result, &CommentInfo{
			ID:        int(c.GetID()),
			Body:      c.GetBody(),
			User:      c.GetUser().GetLogin(),
			CreatedAt: c.GetCreatedAt().String(),
			UpdatedAt: c.GetUpdatedAt().String(),
		})
	}
	return result, nil
}

// Search 搜索 Issues
func (s *IssueService) Search(ctx context.Context, query string, opts *github.SearchOptions) ([]*IssueInfo, int, error) {
	result, _, err := s.client.gh.Search.Issues(ctx, query, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("search issues: %w", err)
	}
	var issues []*IssueInfo
	for _, issue := range result.Issues {
		if issue.IsPullRequest() {
			continue
		}
		issues = append(issues, issueToInfo(issue))
	}
	return issues, result.GetTotal(), nil
}

func issueToInfo(issue *github.Issue) *IssueInfo {
	labels := make([]string, len(issue.Labels))
	for i, l := range issue.Labels {
		labels[i] = l.GetName()
	}
	assignees := make([]string, len(issue.Assignees))
	for i, a := range issue.Assignees {
		assignees[i] = a.GetLogin()
	}
	return &IssueInfo{
		Number:    issue.GetNumber(),
		Title:     issue.GetTitle(),
		State:     issue.GetState(),
		Body:      issue.GetBody(),
		User:      issue.GetUser().GetLogin(),
		Labels:    labels,
		Assignees: assignees,
		URL:       issue.GetHTMLURL(),
		CreatedAt: issue.GetCreatedAt().String(),
		UpdatedAt: issue.GetUpdatedAt().String(),
		Comments:  issue.GetComments(),
	}
}

// FormatIssue 格式化 Issue 为 Markdown
func FormatIssue(issue *IssueInfo) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# #%d %s\n\n", issue.Number, issue.Title))
	sb.WriteString(fmt.Sprintf("- **状态:** %s\n", issue.State))
	sb.WriteString(fmt.Sprintf("- **作者:** %s\n", issue.User))
	sb.WriteString(fmt.Sprintf("- **创建时间:** %s\n", issue.CreatedAt))
	sb.WriteString(fmt.Sprintf("- **评论数:** %d\n", issue.Comments))
	sb.WriteString(fmt.Sprintf("- **链接:** %s\n", issue.URL))
	if len(issue.Labels) > 0 {
		sb.WriteString(fmt.Sprintf("- **标签:** %s\n", strings.Join(issue.Labels, ", ")))
	}
	if len(issue.Assignees) > 0 {
		sb.WriteString(fmt.Sprintf("- **指派:** %s\n", strings.Join(issue.Assignees, ", ")))
	}
	if issue.Body != "" {
		sb.WriteString(fmt.Sprintf("\n## 描述\n\n%s\n", issue.Body))
	}
	return sb.String()
}

// FormatIssueList 格式化 Issue 列表
func FormatIssueList(issues []*IssueInfo, owner, repo string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📋 Issues: %s/%s\n\n", owner, repo))
	if len(issues) == 0 {
		sb.WriteString("无 Issues\n")
		return sb.String()
	}
	for _, issue := range issues {
		state := "🟢"
		if issue.State == "closed" {
			state = "🔴"
		}
		labels := ""
		if len(issue.Labels) > 0 {
			labels = " [" + strings.Join(issue.Labels, ", ") + "]"
		}
		sb.WriteString(fmt.Sprintf("%s #%d %s%s\n", state, issue.Number, issue.Title, labels))
	}
	return sb.String()
}

// FormatCommentList 格式化评论列表
func FormatCommentList(comments []*CommentInfo, issueNum int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("💬 评论列表: #%d\n\n", issueNum))
	if len(comments) == 0 {
		sb.WriteString("无评论\n")
		return sb.String()
	}
	for _, c := range comments {
		sb.WriteString(fmt.Sprintf("---\n**%s** (%s)\n\n%s\n\n", c.User, c.CreatedAt, c.Body))
	}
	return sb.String()
}
