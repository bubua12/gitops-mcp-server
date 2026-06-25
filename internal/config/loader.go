package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load 加载配置，优先级：环境变量 > 配置文件 > 默认值
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath != "" {
		if err := loadFromFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Transport: "stdio",
			Port:      "18080",
		},
		GitHub: GitHubConfig{
			DefaultOwner: "",
		},
		Cache: CacheConfig{
			Backend: "memory",
			TTL: CacheTTLConfig{
				RepoStructure: "1h",
				FileContent:   "5m",
				IssueList:     "2m",
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// loadFromFile 从 YAML 文件加载配置
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}
	return nil
}

// applyEnvOverrides 应用环境变量覆盖
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("GITHUB_TOKEN"); v != "" {
		cfg.GitHub.Token = v
	}
	if v := os.Getenv("GITHUB_DEFAULT_OWNER"); v != "" {
		cfg.GitHub.DefaultOwner = v
	}
	if v := os.Getenv("MCP_TRANSPORT"); v != "" {
		cfg.Server.Transport = v
	}
	if v := os.Getenv("MCP_PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("MCP_API_KEY"); v != "" {
		cfg.Server.APIKey = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}

	// 处理多 Token 配置，格式：name1:token1,name2:token2
	if v := os.Getenv("GITHUB_TOKENS"); v != "" {
		cfg.GitHub.Tokens = parseTokens(v)
	}
}

// ResolveConfigPath 确定配置文件路径
// 优先级：显式传入 > MCP_CONFIG 环境变量 > 默认路径 configs/config.yaml
// 返回空字符串表示没有找到配置文件（走纯环境变量模式）
func ResolveConfigPath(explicit string) string {
	// 1. 显式指定（命令行参数）
	if explicit != "" {
		return explicit
	}

	// 2. 环境变量
	if v := os.Getenv("MCP_CONFIG"); v != "" {
		return v
	}

	// 3. 默认路径：当前目录下的 configs/config.yaml
	defaultPath := filepath.Join("configs", "config.yaml")
	if _, err := os.Stat(defaultPath); err == nil {
		return defaultPath
	}

	// 没有配置文件，走纯环境变量模式
	return ""
}

// parseTokens 解析多 Token 配置
func parseTokens(s string) []TokenConfig {
	var tokens []TokenConfig
	for i, pair := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			tokens = append(tokens, TokenConfig{
				Name:    strings.TrimSpace(parts[0]),
				Token:   strings.TrimSpace(parts[1]),
				Default: i == 0,
			})
		}
	}
	return tokens
}
