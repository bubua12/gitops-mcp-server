package config

// Config 应用配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	GitHub GitHubConfig `yaml:"github"`
	Cache  CacheConfig  `yaml:"cache"`
	Log    LogConfig    `yaml:"log"`
}

// ServerConfig MCP Server 配置
type ServerConfig struct {
	Transport string `yaml:"transport"` // stdio | sse | streamable-http
	Port      string `yaml:"port"`
	APIKey    string `yaml:"api_key"`
}

// GitHubConfig GitHub 相关配置
type GitHubConfig struct {
	Token        string              `yaml:"token"`
	DefaultOwner string              `yaml:"default_owner"`
	Tokens       []TokenConfig       `yaml:"tokens"`
	WatchedRepos []WatchedRepoConfig `yaml:"watched_repos"`
}

// TokenConfig GitHub Token 配置
type TokenConfig struct {
	Name    string `yaml:"name"`
	Token   string `yaml:"token"`
	Default bool   `yaml:"default"`
}

// WatchedRepoConfig 监控仓库配置
type WatchedRepoConfig struct {
	Owner         string `yaml:"owner"`
	Repo          string `yaml:"repo"`
	WatchIssues   bool   `yaml:"watch_issues"`
	WatchReleases bool   `yaml:"watch_releases"`
	WatchSecurity bool   `yaml:"watch_security"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	Backend string        `yaml:"backend"` // memory | redis
	TTL     CacheTTLConfig `yaml:"ttl"`
}

// CacheTTLConfig 缓存 TTL 配置
type CacheTTLConfig struct {
	RepoStructure string `yaml:"repo_structure"`
	FileContent   string `yaml:"file_content"`
	IssueList     string `yaml:"issue_list"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`  // debug | info | warn | error
	Format string `yaml:"format"` // json | text
}
