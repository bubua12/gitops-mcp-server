package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Notification 通知消息
type Notification struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Body      string            `json:"body"`
	Level     string            `json:"level"` // info, warn, error
	Timestamp time.Time         `json:"timestamp"`
	Source    *EventSource      `json:"source,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// EventSource 事件来源
type EventSource struct {
	Type string `json:"type"` // issue_comment, issue_closed, new_release, ci_failure, etc.
	Repo string `json:"repo"`
	// Number int    `json:"number,omitempty"`
}

// Channel 通知渠道接口
type Channel interface {
	Name() string
	Type() string
	Send(ctx context.Context, msg Notification) error
	Test(ctx context.Context) error
}

// Manager 通知管理器
type Manager struct {
	mu       sync.RWMutex
	channels map[string]Channel
}

// NewManager 创建通知管理器
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]Channel),
	}
}

// Register 注册通知渠道
func (m *Manager) Register(ch Channel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[ch.Name()] = ch
}

// Send 发送通知到指定渠道
func (m *Manager) Send(ctx context.Context, channelName string, msg Notification) error {
	m.mu.RLock()
	ch, ok := m.channels[channelName]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("notification channel not found: %s", channelName)
	}
	return ch.Send(ctx, msg)
}

// SendAll 发送通知到所有渠道
func (m *Manager) SendAll(ctx context.Context, msg Notification) []error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errs []error
	for _, ch := range m.channels {
		if err := ch.Send(ctx, msg); err != nil {
			errs = append(errs, fmt.Errorf("channel %s: %w", ch.Name(), err))
		}
	}
	return errs
}

// ListChannels 列出所有渠道
func (m *Manager) ListChannels() []ChannelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ChannelInfo
	for _, ch := range m.channels {
		result = append(result, ChannelInfo{
			Name: ch.Name(),
			Type: ch.Type(),
		})
	}
	return result
}

// TestChannel 测试渠道
func (m *Manager) TestChannel(ctx context.Context, name string) error {
	m.mu.RLock()
	ch, ok := m.channels[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("channel not found: %s", name)
	}
	return ch.Test(ctx)
}

// ChannelInfo 渠道信息
type ChannelInfo struct {
	Name string
	Type string
}

// ChType 返回渠道类型（避免与 Name 方法冲突）
func (m *Manager) ChType() string { return "manager" }

// ========== Terminal 渠道 ==========

// TerminalChannel 终端输出渠道
type TerminalChannel struct {
	name string
}

// NewTerminalChannel 创建终端渠道
func NewTerminalChannel(name string) *TerminalChannel {
	return &TerminalChannel{name: name}
}

func (c *TerminalChannel) Name() string { return c.name }
func (c *TerminalChannel) ChType() string { return "terminal" }
func (c *TerminalChannel) Type() string { return "terminal" }

func (c *TerminalChannel) Send(ctx context.Context, msg Notification) error {
	log.Printf("[NOTIFY] %s [%s] %s\n%s\n", msg.Level, msg.Timestamp.Format(time.RFC3339), msg.Title, msg.Body)
	return nil
}

func (c *TerminalChannel) Test(ctx context.Context) error {
	log.Printf("[NOTIFY TEST] Terminal channel '%s' is working\n", c.name)
	return nil
}

// ========== Webhook 渠道 ==========

// WebhookChannel Webhook 通知渠道
type WebhookChannel struct {
	name    string
	url     string
	client  *http.Client
	headers map[string]string
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	Name    string
	URL     string
	Headers map[string]string
}

// NewWebhookChannel 创建 Webhook 渠道
func NewWebhookChannel(cfg WebhookConfig) *WebhookChannel {
	return &WebhookChannel{
		name:    cfg.Name,
		url:     cfg.URL,
		client:  &http.Client{Timeout: 10 * time.Second},
		headers: cfg.Headers,
	}
}

func (c *WebhookChannel) Name() string { return c.name }
func (c *WebhookChannel) ChType() string { return "webhook" }
func (c *WebhookChannel) Type() string { return "webhook" }

func (c *WebhookChannel) Send(ctx context.Context, msg Notification) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal notification: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (c *WebhookChannel) Test(ctx context.Context) error {
	testMsg := Notification{
		ID:        "test",
		Title:     "GitOps MCP Server 测试通知",
		Body:      "这是一条测试通知，用于验证 Webhook 渠道是否正常工作",
		Level:     "info",
		Timestamp: time.Now(),
	}
	return c.Send(ctx, testMsg)
}
