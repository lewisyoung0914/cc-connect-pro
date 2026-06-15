package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/chenhg5/cc-connect/config"
	"github.com/chenhg5/cc-connect/core"
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
	ctx     context.Context
	service *Service
	status  ServiceStatus
	cfg     *config.Config
	cfgPath string
}

// ServiceStartup is the Wails lifecycle callback invoked when the application starts.
// It implements the application.ServiceStartup interface.
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	a.status = StatusIdle
	a.service = NewService(a)
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
