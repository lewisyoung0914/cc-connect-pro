package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/chenhg5/cc-connect/config"
	"github.com/chenhg5/cc-connect/core"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// ServiceStatus represents the current state of the cc-connect service.
type ServiceStatus string

const (
	StatusIdle     ServiceStatus = "idle"
	StatusStarting ServiceStatus = "starting"
	StatusRunning  ServiceStatus = "running"
	StatusStopping ServiceStatus = "stopping"
	StatusError    ServiceStatus = "error"
)

// ConfigSummary provides a high-level overview of the loaded configuration.
type ConfigSummary struct {
	ConfigPath string        `json:"configPath"`
	DataDir    string        `json:"dataDir"`
	Language   string        `json:"language"`
	Projects   []ProjectInfo `json:"projects"`
}

// ProjectInfo describes a single project from the configuration.
type ProjectInfo struct {
	Name      string `json:"name"`
	AgentType string `json:"agentType"`
	WorkDir   string `json:"workDir"`
	HasFeishu bool   `json:"hasFeishu"`
}

// App is the Wails binding struct exposed to the frontend.
type App struct {
	ctx        context.Context
	service    *Service
	status     ServiceStatus
	cfg        *config.Config
	cfgPath    string
	tray       *application.SystemTray
	trayMenu   *application.Menu
	statusItem *application.MenuItem
	window     *application.WebviewWindow
}

// ServiceStartup is the Wails lifecycle callback invoked when the application starts.
// It implements the application.ServiceStartup interface.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	a.status = StatusIdle
	a.service = NewService(a)

	// Listen for service status events to update the tray
	application.Get().Event.On("service:starting", func(event *application.CustomEvent) {
		a.status = StatusStarting
		if a.tray != nil && a.trayMenu != nil {
			updateTrayStatus(a.tray, a.trayMenu, a.status)
		}
	})
	application.Get().Event.On("service:running", func(event *application.CustomEvent) {
		a.status = StatusRunning
		if a.tray != nil && a.trayMenu != nil {
			updateTrayStatus(a.tray, a.trayMenu, a.status)
		}
	})
	application.Get().Event.On("service:stopping", func(event *application.CustomEvent) {
		a.status = StatusStopping
		if a.tray != nil && a.trayMenu != nil {
			updateTrayStatus(a.tray, a.trayMenu, a.status)
		}
	})
	application.Get().Event.On("service:idle", func(event *application.CustomEvent) {
		a.status = StatusIdle
		if a.tray != nil && a.trayMenu != nil {
			updateTrayStatus(a.tray, a.trayMenu, a.status)
		}
	})
	application.Get().Event.On("service:error", func(event *application.CustomEvent) {
		a.status = StatusError
		if a.tray != nil && a.trayMenu != nil {
			updateTrayStatus(a.tray, a.trayMenu, a.status)
		}
	})

	return nil
}

// Shutdown performs a graceful shutdown of the service.
func (a *App) Shutdown() error {
	if a.service != nil {
		return a.service.Stop()
	}
	return nil
}

// GetServiceStatus returns the current service status.
func (a *App) GetServiceStatus() ServiceStatus {
	return a.status
}

// GetConfigSummary returns a summary of the loaded configuration.
func (a *App) GetConfigSummary() (*ConfigSummary, error) {
	if a.cfg == nil {
		cfgPath := resolveConfigPath(a.cfgPath)
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return nil, err
		}
		a.cfg = cfg
		a.cfgPath = cfgPath
	}

	summary := &ConfigSummary{
		ConfigPath: a.cfgPath,
		DataDir:    a.cfg.DataDir,
		Language:   a.cfg.Language,
		Projects:   make([]ProjectInfo, 0, len(a.cfg.Projects)),
	}

	for _, proj := range a.cfg.Projects {
		workDir, _ := proj.Agent.Options["work_dir"].(string)
		hasFeishu := false
		for _, pc := range proj.Platforms {
			if pc.Type == "feishu" {
				hasFeishu = true
				break
			}
		}
		summary.Projects = append(summary.Projects, ProjectInfo{
			Name:      proj.Name,
			AgentType: proj.Agent.Type,
			WorkDir:   workDir,
			HasFeishu: hasFeishu,
		})
	}

	return summary, nil
}

// HasConfig checks if a config.toml file exists at the resolved path.
func (a *App) HasConfig() bool {
	cfgPath := resolveConfigPath(a.cfgPath)
	return fileExists(cfgPath)
}

// ListRegisteredAgents returns all agent types registered via plugin imports.
func (a *App) ListRegisteredAgents() []string {
	return core.ListRegisteredAgents()
}

// ListRegisteredPlatforms returns all platform types registered via plugin imports.
func (a *App) ListRegisteredPlatforms() []string {
	return core.ListRegisteredPlatforms()
}

// CreateProjectWithFeishuOpts holds the parameters for creating a project with Feishu credentials.
type CreateProjectWithFeishuOpts struct {
	ProjectName string `json:"projectName"`
	AgentType   string `json:"agentType"`
	AppID       string `json:"appId"`
	AppSecret   string `json:"appSecret"`
	WorkDir     string `json:"workDir"`
}

// CreateProjectWithFeishu creates a new project with Feishu platform credentials.
func (a *App) CreateProjectWithFeishu(opts CreateProjectWithFeishuOpts) error {
	if opts.ProjectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if opts.AppID == "" || opts.AppSecret == "" {
		return fmt.Errorf("飞书 App ID 和 App Secret 不能为空")
	}

	dataDir := filepath.Join(homeDir(), ".cc-connect")
	os.MkdirAll(dataDir, 0o755)

	cfgPath := filepath.Join(dataDir, "config.toml")
	if !fileExists(cfgPath) {
		if err := os.WriteFile(cfgPath, []byte(""), 0o644); err != nil {
			return fmt.Errorf("创建配置文件失败: %w", err)
		}
	}

	config.ConfigPath = cfgPath
	a.cfgPath = cfgPath

	result, err := config.EnsureProjectWithFeishuPlatform(config.EnsureProjectWithFeishuOptions{
		ProjectName:  opts.ProjectName,
		PlatformType: "feishu",
		AgentType:    opts.AgentType,
		WorkDir:      opts.WorkDir,
	})
	if err != nil {
		return fmt.Errorf("创建项目失败: %w", err)
	}

	_, err = config.SaveFeishuPlatformCredentials(config.FeishuCredentialUpdateOptions{
		ProjectName:   opts.ProjectName,
		PlatformIndex: 0,
		PlatformType:  "feishu",
		AppID:         opts.AppID,
		AppSecret:     opts.AppSecret,
	})
	if err != nil {
		return fmt.Errorf("写入飞书凭证失败: %w", err)
	}

	slog.Info("项目创建成功", "project", opts.ProjectName, "created", result.Created)
	return nil
}

// StartService starts the cc-connect engine service.
func (a *App) StartService() error {
	err := a.service.Start()
	if err != nil {
		a.status = StatusError
	} else {
		a.status = StatusRunning
	}
	return err
}

// StopService stops the cc-connect engine service.
func (a *App) StopService() error {
	err := a.service.Stop()
	if err != nil {
		a.status = StatusError
	} else {
		a.status = StatusIdle
	}
	return err
}

// RestartService restarts the cc-connect engine service.
func (a *App) RestartService() error {
	err := a.service.Restart()
	if err != nil {
		a.status = StatusError
	} else {
		a.status = StatusRunning
	}
	return err
}

// resolveConfigPath determines which config file to use.
// Priority: explicit path > ./config.toml > ~/.cc-connect/config.toml > ""
func resolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if _, err := os.Stat("config.toml"); err == nil {
		return "config.toml"
	}
	if home := homeDir(); home != "" {
		return filepath.Join(home, ".cc-connect", "config.toml")
	}
	return ""
}

// fileExists checks whether a file exists at the given path.
func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

// homeDir returns the user's home directory, or "" on error.
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// FeishuPlatformDetail contains all feishu/lark platform option fields for a project.
type FeishuPlatformDetail struct {
	ProjectName   string `json:"projectName"`
	PlatformType  string `json:"platformType"`
	AppID         string `json:"appId"`
	AppSecret     string `json:"appSecret"`
	Domain        string `json:"domain"`
	AllowFrom     string `json:"allowFrom"`
	AllowChat     string `json:"allowChat"`
	GroupOnly     bool   `json:"groupOnly"`
	GroupReplyAll bool   `json:"groupReplyAll"`
	ShareSession  bool   `json:"shareSession"`
	ThreadIsolation bool `json:"threadIsolation"`
	ReactionEmoji string `json:"reactionEmoji"`
	DoneEmoji     string `json:"doneEmoji"`
	ProgressStyle string `json:"progressStyle"`
	EnableCard    bool   `json:"enableCard"`
	ResolveMentions bool `json:"resolveMentions"`
	Port          string `json:"port"`
	CallbackPath  string `json:"callbackPath"`
	EncryptKey    string `json:"encryptKey"`
}

// SaveFeishuConfigOpts holds the fields to save back for a feishu/lark platform.
type SaveFeishuConfigOpts struct {
	ProjectName     string `json:"projectName"`
	AppID           string `json:"appId"`
	AppSecret       string `json:"appSecret"`
	Domain          string `json:"domain"`
	AllowFrom       string `json:"allowFrom"`
	AllowChat       string `json:"allowChat"`
	GroupOnly       bool   `json:"groupOnly"`
	GroupReplyAll   bool   `json:"groupReplyAll"`
	ShareSession    bool   `json:"shareSession"`
	ThreadIsolation bool   `json:"threadIsolation"`
	ReactionEmoji   string `json:"reactionEmoji"`
	DoneEmoji       string `json:"doneEmoji"`
	ProgressStyle   string `json:"progressStyle"`
	EnableCard      bool   `json:"enableCard"`
	ResolveMentions bool   `json:"resolveMentions"`
	Port            string `json:"port"`
	CallbackPath    string `json:"callbackPath"`
	EncryptKey      string `json:"encryptKey"`
}

// loadConfig ensures the config is loaded into a.cfg, refreshing if needed.
func (a *App) loadConfig() error {
	if a.cfg == nil {
		cfgPath := resolveConfigPath(a.cfgPath)
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}
		a.cfg = cfg
		a.cfgPath = cfgPath
	}
	return nil
}

// findFeishuPlatform returns the project config and the absolute platform index
// for the first feishu/lark platform in the named project.
func (a *App) findFeishuPlatform(projectName string) (*config.ProjectConfig, int, error) {
	if err := a.loadConfig(); err != nil {
		return nil, -1, err
	}
	for i := range a.cfg.Projects {
		if a.cfg.Projects[i].Name == projectName {
			for j := range a.cfg.Projects[i].Platforms {
				t := strings.ToLower(strings.TrimSpace(a.cfg.Projects[i].Platforms[j].Type))
				if t == "feishu" || t == "lark" {
					return &a.cfg.Projects[i], j, nil
				}
			}
			return nil, -1, fmt.Errorf("项目 %q 没有飞书/Lark 平台", projectName)
		}
	}
	return nil, -1, fmt.Errorf("项目 %q 不存在", projectName)
}

// GetFeishuConfigDetail returns all feishu/lark platform option fields for a project.
func (a *App) GetFeishuConfigDetail(projectName string) (*FeishuPlatformDetail, error) {
	proj, platformIdx, err := a.findFeishuPlatform(projectName)
	if err != nil {
		return nil, err
	}
	platform := proj.Platforms[platformIdx]
	opts := platform.Options

	detail := &FeishuPlatformDetail{
		ProjectName:  projectName,
		PlatformType: platform.Type,
		AppID:        stringOpt(opts, "app_id"),
		AppSecret:    stringOpt(opts, "app_secret"),
		Domain:       stringOpt(opts, "domain"),
		AllowFrom:    stringOpt(opts, "allow_from"),
		AllowChat:    stringOpt(opts, "allow_chat"),
		GroupOnly:    boolOpt(opts, "group_only"),
		GroupReplyAll: boolOpt(opts, "group_reply_all"),
		ShareSession:  boolOpt(opts, "share_session_in_channel"),
		ThreadIsolation: boolOpt(opts, "thread_isolation"),
		ReactionEmoji: stringOpt(opts, "reaction_emoji"),
		DoneEmoji:     stringOpt(opts, "done_emoji"),
		ProgressStyle: stringOpt(opts, "progress_style"),
		EnableCard:    boolOpt(opts, "enable_feishu_card"),
		ResolveMentions: boolOpt(opts, "resolve_mentions"),
		Port:          stringOpt(opts, "port"),
		CallbackPath:  stringOpt(opts, "callback_path"),
		EncryptKey:    stringOpt(opts, "encrypt_key"),
	}

	return detail, nil
}

// SaveFeishuConfig saves feishu/lark platform configuration back to config.toml.
func (a *App) SaveFeishuConfig(opts SaveFeishuConfigOpts) error {
	if err := a.loadConfig(); err != nil {
		return err
	}

	// Find project and platform
	_, platformIdx, err := a.findFeishuPlatform(opts.ProjectName)
	if err != nil {
		return err
	}

	// Set ConfigPath for the config package functions
	config.ConfigPath = a.cfgPath

	// Step 1: Save app_id/app_secret via SaveFeishuPlatformCredentials
	_, err = config.SaveFeishuPlatformCredentials(config.FeishuCredentialUpdateOptions{
		ProjectName:   opts.ProjectName,
		PlatformIndex: 0,
		AppID:         opts.AppID,
		AppSecret:     opts.AppSecret,
	})
	if err != nil {
		return fmt.Errorf("写入飞书凭证失败: %w", err)
	}

	// Step 2: Save other platform options via SaveFeishuPlatformOptions
	platformOpts := []config.PlatformOptionUpdate{
		{Key: "domain", Value: opts.Domain},
		{Key: "allow_from", Value: opts.AllowFrom},
		{Key: "allow_chat", Value: opts.AllowChat},
		{Key: "group_only", Value: opts.GroupOnly},
		{Key: "group_reply_all", Value: opts.GroupReplyAll},
		{Key: "share_session_in_channel", Value: opts.ShareSession},
		{Key: "thread_isolation", Value: opts.ThreadIsolation},
		{Key: "reaction_emoji", Value: opts.ReactionEmoji},
		{Key: "done_emoji", Value: opts.DoneEmoji},
		{Key: "progress_style", Value: opts.ProgressStyle},
		{Key: "enable_feishu_card", Value: opts.EnableCard},
		{Key: "resolve_mentions", Value: opts.ResolveMentions},
		{Key: "port", Value: opts.Port},
		{Key: "callback_path", Value: opts.CallbackPath},
		{Key: "encrypt_key", Value: opts.EncryptKey},
	}

	// Only save non-empty/changed options
	filteredOpts := platformOpts[:0]
	for _, opt := range platformOpts {
		if v, ok := opt.Value.(string); ok && v == "" {
			continue // skip empty string options (they weren't set)
		}
		filteredOpts = append(filteredOpts, opt)
	}

	if len(filteredOpts) > 0 {
		err = config.SaveFeishuPlatformOptions(opts.ProjectName, platformIdx, filteredOpts)
		if err != nil {
			return fmt.Errorf("写入平台选项失败: %w", err)
		}
	}

	// Reload config in memory
	a.cfg, err = config.Load(a.cfgPath)
	if err != nil {
		return fmt.Errorf("重载配置失败: %w", err)
	}

	// If service is running, restart to pick up changes
	if a.status == StatusRunning {
		slog.Info("服务运行中，重启以应用配置变更")
		if err := a.service.Restart(); err != nil {
			return fmt.Errorf("重启服务失败: %w", err)
		}
	}

	return nil
}

// ValidateFeishuCredentials verifies that the given app_id, app_secret, and domain
// are valid by attempting to obtain a tenant access token and fetch bot info.
func (a *App) ValidateFeishuCredentials(appID, appSecret, domain string) error {
	if appID == "" || appSecret == "" {
		return fmt.Errorf("App ID 和 App Secret 不能为空")
	}

	// Determine domain
	domain = strings.TrimSpace(domain)
	if domain == "" {
		domain = "https://open.feishu.cn"
	}

	// Create a temporary lark client with the provided credentials
	var clientOpts []lark.ClientOptionFunc
	if domain != lark.FeishuBaseUrl {
		clientOpts = append(clientOpts, lark.WithOpenBaseUrl(domain))
	}
	client := lark.NewClient(appID, appSecret, clientOpts...)

	// Try to get tenant info to verify credentials
	ctx := context.Background()
	resp, err := client.Tenant.Tenant.Query(ctx)
	if err != nil {
		return fmt.Errorf("凭证验证失败: %w", err)
	}

	if !resp.Success() {
		return fmt.Errorf("凭证验证失败: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	return nil
}

// stringOpt extracts a string value from a map[string]any, returning "" if missing.
func stringOpt(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// boolOpt extracts a boolean value from a map[string]any, returning false if missing.
func boolOpt(m map[string]any, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
