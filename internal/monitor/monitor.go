package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gitops-mcp-server/internal/github"
	"gitops-mcp-server/internal/notify"
)

// Monitor 监控器
type Monitor struct {
	Name       string
	Repos      []RepoRef
	Filters    []Filter
	Interval   time.Duration
	NotifyTo   []string
	LastCheck  time.Time
	LastEvents []Event
	Running    bool
}

// RepoRef 仓库引用
type RepoRef struct {
	Owner string
	Repo  string
}

// Filter 过滤器
type Filter struct {
	Type      string // issue_comment, issue_closed, new_release, ci_failure, etc.
	Condition string // 可选的条件表达式
}

// Event 监控事件
type Event struct {
	ID        string
	Type      string
	Repo      string
	Title     string
	Body      string
	URL       string
	Actor     string
	Timestamp time.Time
}

// Engine 监控引擎
type Engine struct {
	mu        sync.RWMutex
	monitors  map[string]*Monitor
	ghClient  *github.Client
	notifyMgr *notify.Manager
	cancel    context.CancelFunc
	events    []Event // 事件日志
}

// NewEngine 创建监控引擎
func NewEngine(ghClient *github.Client, notifyMgr *notify.Manager) *Engine {
	return &Engine{
		monitors:  make(map[string]*Monitor),
		ghClient:  ghClient,
		notifyMgr: notifyMgr,
	}
}

// AddMonitor 添加监控规则
func (e *Engine) AddMonitor(m *Monitor) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.monitors[m.Name]; exists {
		return fmt.Errorf("monitor '%s' already exists", m.Name)
	}

	e.monitors[m.Name] = m
	return nil
}

// RemoveMonitor 移除监控规则
func (e *Engine) RemoveMonitor(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.monitors[name]; !exists {
		return fmt.Errorf("monitor '%s' not found", name)
	}

	delete(e.monitors, name)
	return nil
}

// PauseMonitor 暂停监控
func (e *Engine) PauseMonitor(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	m, exists := e.monitors[name]
	if !exists {
		return fmt.Errorf("monitor '%s' not found", name)
	}

	m.Running = false
	return nil
}

// ResumeMonitor 恢复监控
func (e *Engine) ResumeMonitor(name string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	m, exists := e.monitors[name]
	if !exists {
		return fmt.Errorf("monitor '%s' not found", name)
	}

	m.Running = true
	return nil
}

// ListMonitors 列出所有监控规则
func (e *Engine) ListMonitors() []*Monitor {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*Monitor
	for _, m := range e.monitors {
		result = append(result, m)
	}
	return result
}

// GetEvents 获取事件日志
func (e *Engine) GetEvents(since time.Time, limit int) []Event {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []Event
	for i := len(e.events) - 1; i >= 0; i-- {
		if len(result) >= limit {
			break
		}
		if e.events[i].Timestamp.After(since) {
			result = append(result, e.events[i])
		}
	}
	return result
}

// Start 启动监控引擎
func (e *Engine) Start(ctx context.Context) {
	ctx, e.cancel = context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				e.runChecks(ctx)
			}
		}
	}()

	log.Println("Monitor engine started")
}

// Stop 停止监控引擎
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
	log.Println("Monitor engine stopped")
}

// runChecks 运行所有监控检查
func (e *Engine) runChecks(ctx context.Context) {
	e.mu.RLock()
	var monitors []*Monitor
	for _, m := range e.monitors {
		if m.Running {
			monitors = append(monitors, m)
		}
	}
	e.mu.RUnlock()

	for _, m := range monitors {
		if time.Since(m.LastCheck) < m.Interval {
			continue
		}

		e.checkMonitor(ctx, m)
		m.LastCheck = time.Now()
	}
}

// checkMonitor 检查单个监控规则
func (e *Engine) checkMonitor(ctx context.Context, m *Monitor) {
	for _, repo := range m.Repos {
		// 检查新的 Issue 评论
		if e.shouldFilter(m, "issue_comment") {
			e.checkIssueComments(ctx, m, repo)
		}

		// 检查新的 Release
		if e.shouldFilter(m, "new_release") {
			e.checkNewReleases(ctx, m, repo)
		}

		// 检查 CI 失败
		if e.shouldFilter(m, "ci_failure") {
			e.checkCIFailures(ctx, m, repo)
		}
	}
}

// shouldFilter 检查是否应该过滤
func (e *Engine) shouldFilter(m *Monitor, eventType string) bool {
	if len(m.Filters) == 0 {
		return true // 无过滤器则全部检查
	}
	for _, f := range m.Filters {
		if f.Type == eventType {
			return true
		}
	}
	return false
}

// checkIssueComments 检查 Issue 评论
func (e *Engine) checkIssueComments(ctx context.Context, m *Monitor, repo RepoRef) {
	issues, err := e.ghClient.Issues().List(ctx, repo.Owner, repo.Repo, nil)
	if err != nil {
		log.Printf("Monitor '%s': failed to list issues for %s/%s: %v", m.Name, repo.Owner, repo.Repo, err)
		return
	}

	for _, issue := range issues {
		since := m.LastCheck
		comments, err := e.ghClient.Issues().ListComments(ctx, repo.Owner, repo.Repo, issue.Number, &since)
		if err != nil {
			continue
		}

		for _, comment := range comments {
			event := Event{
				ID:        fmt.Sprintf("comment-%s/%s#%d-%d", repo.Owner, repo.Repo, issue.Number, comment.ID),
				Type:      "issue_comment",
				Repo:      fmt.Sprintf("%s/%s", repo.Owner, repo.Repo),
				Title:     fmt.Sprintf("新评论: %s/#%d", repo.Repo, issue.Number),
				Body:      fmt.Sprintf("%s 评论了 Issue #%d: %s", comment.User, issue.Number, truncate(comment.Body, 200)),
				URL:       fmt.Sprintf("https://github.com/%s/%s/issues/%d", repo.Owner, repo.Repo, issue.Number),
				Actor:     comment.User,
				Timestamp: time.Now(),
			}

			e.addEvent(event)
			e.notifyMonitor(m, event)
		}
	}
}

// checkNewReleases 检查新 Release
func (e *Engine) checkNewReleases(ctx context.Context, m *Monitor, repo RepoRef) {
	releases, err := e.ghClient.Releases().ListReleases(ctx, repo.Owner, repo.Repo, nil)
	if err != nil {
		log.Printf("Monitor '%s': failed to list releases for %s/%s: %v", m.Name, repo.Owner, repo.Repo, err)
		return
	}

	for _, release := range releases {
		// 检查是否是新 Release（在上次检查之后发布的）
		if release.PublishedAt != "" {
			event := Event{
				ID:        fmt.Sprintf("release-%s/%s-%s", repo.Owner, repo.Repo, release.TagName),
				Type:      "new_release",
				Repo:      fmt.Sprintf("%s/%s", repo.Owner, repo.Repo),
				Title:     fmt.Sprintf("新 Release: %s", release.TagName),
				Body:      fmt.Sprintf("%s 发布了 %s", release.Author, release.Name),
				URL:       release.HTMLURL,
				Actor:     release.Author,
				Timestamp: time.Now(),
			}

			e.addEvent(event)
			e.notifyMonitor(m, event)
		}
	}
}

// checkCIFailures 检查 CI 失败
func (e *Engine) checkCIFailures(ctx context.Context, m *Monitor, repo RepoRef) {
	runs, err := e.ghClient.Actions().ListWorkflowRuns(ctx, repo.Owner, repo.Repo, nil)
	if err != nil {
		log.Printf("Monitor '%s': failed to list workflow runs for %s/%s: %v", m.Name, repo.Owner, repo.Repo, err)
		return
	}

	for _, run := range runs {
		if run.Status == "completed" && run.Conclusion == "failure" {
			event := Event{
				ID:        fmt.Sprintf("ci-failure-%s/%s-%d", repo.Owner, repo.Repo, run.ID),
				Type:      "ci_failure",
				Repo:      fmt.Sprintf("%s/%s", repo.Owner, repo.Repo),
				Title:     fmt.Sprintf("CI 失败: %s", run.Name),
				Body:      fmt.Sprintf("工作流 '%s' 在分支 '%s' 运行失败", run.Name, run.Branch),
				URL:       run.HTMLURL,
				Timestamp: time.Now(),
			}

			e.addEvent(event)
			e.notifyMonitor(m, event)
		}
	}
}

// addEvent 添加事件到日志
func (e *Engine) addEvent(event Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.events = append(e.events, event)
	// 保留最近 1000 条事件
	if len(e.events) > 1000 {
		e.events = e.events[len(e.events)-1000:]
	}
}

// notifyMonitor 发送通知
func (e *Engine) notifyMonitor(m *Monitor, event Event) {
	if e.notifyMgr == nil {
		return
	}

	msg := notify.Notification{
		ID:    event.ID,
		Title: event.Title,
		Body:  event.Body,
		Level: "info",
		Source: &notify.EventSource{
			Type: event.Type,
			Repo: event.Repo,
		},
		Timestamp: event.Timestamp,
	}

	if event.Type == "ci_failure" {
		msg.Level = "warn"
	}

	for _, channelName := range m.NotifyTo {
		if err := e.notifyMgr.Send(context.Background(), channelName, msg); err != nil {
			log.Printf("Failed to send notification to '%s': %v", channelName, err)
		}
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// FormatMonitorList 格式化监控列表
func FormatMonitorList(monitors []*Monitor) string {
	result := "📡 监控规则列表\n\n"
	if len(monitors) == 0 {
		result += "无监控规则\n"
		return result
	}
	for _, m := range monitors {
		status := "🟢 运行中"
		if !m.Running {
			status = "⚪ 已暂停"
		}
		result += fmt.Sprintf("- **%s** [%s] 间隔: %s\n", m.Name, status, m.Interval)
		for _, repo := range m.Repos {
			result += fmt.Sprintf("  - %s/%s\n", repo.Owner, repo.Repo)
		}
	}
	return result
}

// FormatEventList 格式化事件列表
func FormatEventList(events []Event) string {
	result := "📋 事件日志\n\n"
	if len(events) == 0 {
		result += "无事件\n"
		return result
	}
	for _, e := range events {
		icon := "📌"
		switch e.Type {
		case "issue_comment":
			icon = "💬"
		case "new_release":
			icon = "🚀"
		case "ci_failure":
			icon = "❌"
		}
		result += fmt.Sprintf("%s [%s] %s\n  %s\n  %s\n\n", icon, e.Timestamp.Format("01-02 15:04"), e.Title, e.Body, e.URL)
	}
	return result
}
