package github

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// Client 封装 GitHub API 客户端
type Client struct {
	gh           *github.Client
	token        string
	defaultOwner string
}

// NewClient 创建 GitHub 客户端
func NewClient(token string, defaultOwner string) (*Client, error) {
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return nil, fmt.Errorf("GitHub token is required, set GITHUB_TOKEN env or pass in config")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	gh := github.NewClient(tc)

	return &Client{
		gh:           gh,
		token:        token,
		defaultOwner: defaultOwner,
	}, nil
}

// Ping 验证 Token 并检查连接
func (c *Client) Ping(ctx context.Context) error {
	user, _, err := c.gh.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("GitHub auth failed: %w", err)
	}
	_ = user
	return nil
}

// GetDefaultOwner 获取默认 owner
func (c *Client) GetDefaultOwner() string {
	return c.defaultOwner
}

// resolveOwner 解析 owner，如果为空则使用 defaultOwner
func (c *Client) resolveOwner(owner string) string {
	if owner == "" {
		return c.defaultOwner
	}
	return owner
}

// CheckRateLimit 检查速率限制状态
func (c *Client) CheckRateLimit() (remaining int, resetAt time.Time, warn bool) {
	ctx := context.Background()
	limits, _, err := c.gh.RateLimit.Get(ctx)
	if err != nil {
		return 0, time.Time{}, false
	}
	if limits.Core == nil {
		return 0, time.Time{}, false
	}
	return limits.Core.Remaining, limits.Core.Reset.Time, limits.Core.Remaining < 500
}

// Repositories 返回仓库相关操作
func (c *Client) Repositories() *RepositoryService {
	return &RepositoryService{client: c}
}

// Issues 返回 Issue 相关操作
func (c *Client) Issues() *IssueService {
	return &IssueService{client: c}
}

// PullRequests 返回 PR 相关操作
func (c *Client) PullRequests() *PullRequestService {
	return &PullRequestService{client: c}
}

// Releases 返回 Release 相关操作
func (c *Client) Releases() *ReleaseService {
	return &ReleaseService{client: c}
}

// Actions 返回 CI/CD 相关操作
func (c *Client) Actions() *ActionsService {
	return &ActionsService{client: c}
}

// Git 返回 Git 相关操作
func (c *Client) Git() *GitService {
	return &GitService{client: c}
}

// RawHTTPClient 返回底层 HTTP Client（用于自定义请求）
func (c *Client) RawHTTPClient() *http.Client {
	return c.gh.Client()
}

// Underlying 返回底层 go-github 客户端
func (c *Client) Underlying() *github.Client {
	return c.gh
}
