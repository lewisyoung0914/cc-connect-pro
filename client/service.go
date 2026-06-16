package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chenhg5/cc-connect/config"
	"github.com/chenhg5/cc-connect/core"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows/registry"
)

// Service manages the cc-connect engine lifecycle.
type Service struct {
	app     *App
	mu      sync.Mutex
	engines []*core.Engine
	status  ServiceStatus
	errMsg  string
	cancel  context.CancelFunc
}

// NewService creates a new Service bound to the given App.
func NewService(app *App) *Service {
	return &Service{
		app:    app,
		status: StatusIdle,
	}
}

// Start loads config, creates engines for each project, and starts them.
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusRunning || s.status == StatusStarting {
		return fmt.Errorf("service is already running (status: %s)", s.status)
	}

	s.status = StatusStarting
	s.errMsg = ""
	emitEvent("service:starting", nil)

	// Inject environment variables needed by agents on Windows.
	// Claude Code requires git-bash; if it's not in PATH, set the
	// CLAUDE_CODE_GIT_BASH_PATH env var so the agent subprocess can find it.
	injectAgentEnvVars()

	cfgPath := resolveConfigPath(s.app.cfgPath)
	if cfgPath == "" {
		s.status = StatusError
		s.errMsg = "no config file found"
		emitEvent("service:error", s.errMsg)
		return errors.New(s.errMsg)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		s.status = StatusError
		s.errMsg = fmt.Sprintf("failed to load config: %v", err)
		emitEvent("service:error", s.errMsg)
		return errors.New(s.errMsg)
	}

	config.ConfigPath = cfgPath
	s.app.cfg = cfg
	s.app.cfgPath = cfgPath

	if len(cfg.Projects) == 0 {
		s.status = StatusError
		s.errMsg = "no projects configured"
		emitEvent("service:error", s.errMsg)
		return errors.New(s.errMsg)
	}

	_, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	s.engines = make([]*core.Engine, 0, len(cfg.Projects))

	for _, proj := range cfg.Projects {
		engine, err := s.createEngine(cfg, proj)
		if err != nil {
			slog.Error("failed to create engine", "project", proj.Name, "error", err)
			// Stop already-created engines before returning
			for _, e := range s.engines {
				e.Stop()
			}
			s.engines = nil
			s.status = StatusError
			s.errMsg = fmt.Sprintf("failed to create engine for project %q: %v", proj.Name, err)
			emitEvent("service:error", s.errMsg)
			return errors.New(s.errMsg)
		}
		s.engines = append(s.engines, engine)
	}

	// Start all engines
	for _, e := range s.engines {
		if err := e.Start(); err != nil {
			slog.Error("failed to start engine", "error", err)
			s.status = StatusError
			s.errMsg = fmt.Sprintf("failed to start engine: %v", err)
			emitEvent("service:error", s.errMsg)
			return errors.New(s.errMsg)
		}
	}

	s.status = StatusRunning
	emitEvent("service:running", nil)
	return nil
}

// Stop stops all engines, cancels the context, and resets state.
// A 150s overall timeout ensures the service:idle event is always emitted,
// even if an engine.Stop() or agentSession.Close() hangs. Without this
// cap, the desktop client UI would stay stuck at "stopping" forever.
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != StatusRunning && s.status != StatusStarting {
		return nil
	}

	s.status = StatusStopping
	emitEvent("service:stopping", nil)

	// Overall timeout — covers the 130s closeAgentSessionWithTimeout per
	// session plus a small buffer. If any engine hangs beyond this, we
	// force-reset state so the UI can recover.
	const stopTimeout = 150 * time.Second
	done := make(chan struct{})
	go func() {
		for _, e := range s.engines {
			if err := e.Stop(); err != nil {
				slog.Warn("engine stop error", "error", err)
			}
		}
		close(done)
	}()

	select {
	case <-done:
		// All engines stopped cleanly.
	case <-time.After(stopTimeout):
		slog.Error("service stop timed out, forcing state reset",
			"timeout", stopTimeout)
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.engines = nil
	s.cancel = nil
	s.errMsg = ""
	s.status = StatusIdle
	emitEvent("service:idle", nil)
	return nil
}

// Restart stops the service and then starts it again.
func (s *Service) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

// GetEngines returns the currently running engines.
func (s *Service) GetEngines() []*core.Engine {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engines
}

// createEngine builds one engine from a project config, following the same
// pattern as cmd/cc-connect/main.go but simplified for the desktop client.
func (s *Service) createEngine(cfg *config.Config, proj config.ProjectConfig) (*core.Engine, error) {
	// Build agent options
	agentOpts := make(map[string]any, len(proj.Agent.Options)+2)
	for k, v := range proj.Agent.Options {
		agentOpts[k] = v
	}
	agentOpts["cc_data_dir"] = cfg.DataDir
	agentOpts["cc_project"] = proj.Name

	agent, err := core.CreateAgent(proj.Agent.Type, agentOpts)
	if err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	// Build platforms
	var platforms []core.Platform
	for _, pc := range proj.Platforms {
		opts := make(map[string]any, len(pc.Options)+3)
		for k, v := range pc.Options {
			opts[k] = v
		}
		opts["cc_data_dir"] = cfg.DataDir
		opts["cc_project"] = proj.Name
		opts["cc_platform_name"] = pc.Name
		p, err := core.CreatePlatform(pc.Type, opts)
		if err != nil {
			return nil, fmt.Errorf("create platform %q: %w", pc.Type, err)
		}
		platforms = append(platforms, p)
	}

	workDir, _ := proj.Agent.Options["work_dir"].(string)
	sessionFile := sessionStorePath(cfg.DataDir, proj.Name, workDir)

	// Parse language setting
	var lang core.Language
	switch cfg.Language {
	case "zh", "chinese":
		lang = core.LangChinese
	case "zh-TW", "zh_TW", "zhtw":
		lang = core.LangTraditionalChinese
	case "ja", "japanese":
		lang = core.LangJapanese
	case "es", "spanish":
		lang = core.LangSpanish
	case "en", "english":
		lang = core.LangEnglish
	default:
		lang = core.LangAuto
	}

	engine := core.NewEngine(proj.Name, agent, platforms, sessionFile, lang)

	// Set basic engine config
	engine.SetDataDir(cfg.DataDir)
	engine.SetAttachmentSendEnabled(cfg.AttachmentSend != "off")
	engine.SetBaseWorkDir(workDir)

	// Wire display config
	{
		mode, tm, tool, tmlen, toollen, showCtx, showFooter := config.EffectiveDisplay(cfg, &proj)
		engine.SetDisplayConfig(core.DisplayCfg{
			Mode:             mode,
			CardMode:         config.EffectiveCardMode(cfg, &proj),
			ThinkingMessages: tm,
			ThinkingMaxLen:   tmlen,
			ToolMaxLen:       toollen,
			ToolMessages:     tool,
		})
		engine.SetShowContextIndicator(showCtx)
		engine.SetReplyFooterEnabled(showFooter)
	}

	// Wire project state store
	projectState := core.NewProjectStateStore(projectStatePath(cfg.DataDir, proj.Name))
	engine.SetProjectStateStore(projectState)

	return engine, nil
}

// emitEvent safely emits a Wails event. It checks that the global
// application instance is available to avoid panics in tests where
// no Wails runtime is running.
func emitEvent(name string, data any) {
	app := application.Get()
	if app == nil {
		return
	}
	app.Event.Emit(name, data)
}

// sessionStorePath returns the path to the session store JSON file.
// Follows the same naming convention as cmd/cc-connect/main.go.
func sessionStorePath(dataDir, name, workDir string) string {
	var filename string
	if workDir == "" {
		filename = name + ".json"
	} else {
		abs, err := filepath.Abs(workDir)
		if err != nil {
			abs = workDir
		}
		h := sha256.Sum256([]byte(abs))
		short := hex.EncodeToString(h[:4])
		filename = fmt.Sprintf("%s-%s.json", name, short)
	}
	return filepath.Join(dataDir, "sessions", filename)
}

// projectStatePath returns the path to the project state file.
// Follows the same naming convention as cmd/cc-connect/main.go.
func projectStatePath(dataDir, projectName string) string {
	replacer := strings.NewReplacer(
		"\\", "_",
		"/", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	safe := replacer.Replace(projectName)
	return filepath.Join(dataDir, "project_state", safe+".json")
}

// injectAgentEnvVars sets environment variables required by agent subprocesses.
// On Windows, Claude Code requires git-bash. If CLAUDE_CODE_GIT_BASH_PATH
// is not already set, this function searches for bash.exe via PATH, well-known
// directories, and the Windows Registry (GitForWindows InstallPath key).
func injectAgentEnvVars() {
	if os.Getenv("CLAUDE_CODE_GIT_BASH_PATH") != "" {
		slog.Info("agent env var already set", "key", "CLAUDE_CODE_GIT_BASH_PATH", "value", os.Getenv("CLAUDE_CODE_GIT_BASH_PATH"))
		return
	}
	if runtime.GOOS != "windows" {
		return
	}

	bashPath := findGitBash()
	if bashPath != "" {
		os.Setenv("CLAUDE_CODE_GIT_BASH_PATH", bashPath)
		slog.Info("injected agent env var", "key", "CLAUDE_CODE_GIT_BASH_PATH", "value", bashPath)
	} else {
		slog.Warn("could not find git-bash; Claude Code agent will fail on Windows")
	}
}

func findGitBash() string {
	// Strategy 1: PATH lookup → derive bash.exe from git.exe location.
	if gitPath, err := exec.LookPath("git.exe"); err == nil {
		gitDir := filepath.Dir(filepath.Dir(gitPath))
		candidate := filepath.Join(gitDir, "bin", "bash.exe")
		if fileExists(candidate) {
			return candidate
		}
	}

	// Strategy 2: Windows Registry — Git for Windows writes InstallPath.
	if installPath := readRegString(registry.LOCAL_MACHINE, "SOFTWARE\\GitForWindows", "InstallPath"); installPath != "" {
		candidate := filepath.Join(installPath, "bin", "bash.exe")
		if fileExists(candidate) {
			return candidate
		}
	}

	// Strategy 3: well-known install directories.
	for _, p := range []string{
		"C:\\Program Files\\Git\\bin\\bash.exe",
		"C:\\Program Files (x86)\\Git\\bin\\bash.exe",
	} {
		if fileExists(p) {
			return p
		}
	}

	return ""
}

func readRegString(k registry.Key, path, name string) string {
	key, err := registry.OpenKey(k, path, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()
	val, _, err := key.GetStringValue(name)
	if err != nil {
		return ""
	}
	return val
}
