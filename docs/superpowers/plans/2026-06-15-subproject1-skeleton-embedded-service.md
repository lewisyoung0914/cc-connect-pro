# 子项目 1：基础骨架 + 内嵌服务 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建 Wails 桌面客户端骨架，内嵌 cc-connect 服务进程，提供侧边栏导航、Dashboard 空壳、视觉设计系统基础、首次启动引导流程。

**Architecture:** 在 cc-connect-pro 项目内新建 `client/` 目录作为独立 Wails v3 子项目，拥有自己的 go.mod（通过 replace 指令引用父模块）。Go 侧直接 import config/daemon/core 包复用代码。前端使用 React + TypeScript + Tailwind，与现有 web/ 技术栈一致。

**Tech Stack:** Wails v3, Go 1.25, React 19, TypeScript, Tailwind CSS 3, Vite 6

---

## 文件结构

```
client/
├── go.mod                          ← 独立 Go module，replace 指令引用父模块
├── go.sum
├── wails.json                      ← Wails v3 项目配置
├── main.go                         ← Wails 入口：初始化 App、绑定、窗口、托盘
├── app.go                          ← App 结构体 + 暴露给前端的方法
├── service.go                      ← 服务状态机 + Engine 启动/停止逻辑
├── service_test.go                 ← service.go 单元测试
├── plugin_platform_feishu.go       ← 空白 import 触发飞书注册
├── plugin_agent_claudecode.go      ← 空白 import 触发 claudecode 注册
├── plugin_agent_codex.go           ← 空白 import 触发 codex 注册
├── frontend/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   ├── tailwind.config.ts          ← 视觉设计系统 token
│   ├── postcss.config.js
│   ├── src/
│   │   ├── main.tsx                ← React 入口
│   │   ├── App.tsx                 ← 主布局：侧边栏 + 内容区 + 路由
│   │   ├── styles/
│   │   │   └── index.css           ← Tailwind 基础 + 自定义样式
│   │   ├── layouts/
│   │   │   └── SidebarLayout.tsx   ← 侧边栏 + 内容区布局组件
│   │   ├── pages/
│   │   │   ├── Dashboard.tsx       ← 服务状态总览页（空壳 + 签名动画）
│   │   │   └── Welcome.tsx         ← 首次启动引导页
│   │   ├── components/
│   │   │   ├── StatusDot.tsx       ← 状态圆点色标组件
│   │   │   ├── SegmentedControl.tsx ← 分段控件组件
│   │   │   └── ConnectionBridge.tsx ← 签名动画：连接桥
│   │   └── hooks/
│   │   │   └ useServiceStatus.ts   ← 监听服务状态的 hook
│   │   └── lib/
│   │   │   └── types.ts            ← 前端共享类型定义
│   │   ├── wailsjs/                ← Wails 自动生成（不走手动）
│   │   └── bindings/               ← Wails 自动生成
│   └── public/
│       └── favicon.ico
├── build/
│   ├── appicon.png                 ← 应用图标
│   └── windows/                    ← Windows 资源（icon.ico 等）
│   └── darwin/                     ← macOS 资源（Info.plist 等）
└── Makefile                        ← 客户端独立构建命令
```

---

### Task 1: 安装 Wails v3 CLI

- [ ] **Step 1: 检查 wails3 是否已安装**

Run: `wails3 version`
Expected: 如果已安装则显示版本号；如果未安装则报错

- [ ] **Step 2: 安装 Wails v3 CLI（如果未安装）**

Run: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
Expected: 安装成功，无报错

- [ ] **Step 3: 验证安装**

Run: `wails3 version`
Expected: 显示 Wails v3 版本号

---

### Task 2: 创建 client/ 目录和 go.mod

**Files:**
- Create: `client/go.mod`
- Create: `client/main.go` (骨架)

- [ ] **Step 1: 创建 client 目录**

Run: `mkdir -p /e/project/cc-connect-pro/client`
Expected: 目录创建成功

- [ ] **Step 2: 创建 go.mod，用 replace 指令引用父模块**

```go
module github.com/chenhg5/cc-connect/client

go 1.25.0

require github.com/chenhg5/cc-connect v0.0.0

replace github.com/chenhg5/cc-connect => ../
```

创建文件 `client/go.mod`

- [ ] **Step 3: 创建 main.go 骨架**

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "cc-connect",
		Description: "cc-connect Desktop Client",
		Assets: application.AssetOptions{
			Handler: application.AssetHandlerFromFS(assets),
		},
	})

	app.Bind(&App{})

	app.NewWindow(&application.WindowOptions{
		Title:  "cc-connect",
		Width:  960,
		Height: 640,
		URL:    "/",
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
```

创建文件 `client/main.go`

- [ ] **Step 4: 添加 Wails 依赖**

Run: `cd /e/project/cc-connect-pro/client && go mod tidy`
Expected: go.mod 和 go.sum 生成，Wails v3 依赖被下载

---

### Task 3: 创建 app.go — App 结构体和基础方法

**Files:**
- Create: `client/app.go`

- [ ] **Step 1: 创建 app.go 骨架**

```go
package main

import (
	"context"
	"fmt"

	"github.com/chenhg5/cc-connect/config"
	"github.com/chenhg5/cc-connect/core"
)

// ServiceStatus 表示服务当前状态
type ServiceStatus string

const (
	StatusIdle    ServiceStatus = "idle"
	StatusStarting ServiceStatus = "starting"
	StatusRunning  ServiceStatus = "running"
	StatusStopping ServiceStatus = "stopping"
	StatusError    ServiceStatus = "error"
)

// ConfigSummary 返回给前端的配置概要
type ConfigSummary struct {
	ConfigPath string        `json:"configPath"`
	DataDir    string        `json:"dataDir"`
	Language   string        `json:"language"`
	Projects   []ProjectInfo `json:"projects"`
}

// ProjectInfo 返回给前端的项目信息
type ProjectInfo struct {
	Name       string `json:"name"`
	AgentType  string `json:"agentType"`
	WorkDir    string `json:"workDir"`
	HasFeishu  bool   `json:"hasFeishu"`
}

// App 是 Wails 绑定的主结构体，所有暴露给前端的方法都在这里
type App struct {
	ctx      context.Context
	service  *Service
	status   ServiceStatus
	cfg      *config.Config
	cfgPath  string
}

// startup 是 Wails 生命周期回调，在窗口创建后调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.service = NewService(a)
	a.cfgPath = resolveConfigPath("")
}

// GetServiceStatus 返回当前服务状态
func (a *App) GetServiceStatus() ServiceStatus {
	return a.status
}

// GetConfigSummary 返回配置概要信息
func (a *App) GetConfigSummary() (*ConfigSummary, error) {
	if a.cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}
	projects := make([]ProjectInfo, 0, len(a.cfg.Projects))
	for _, p := range a.cfg.Projects {
		hasFeishu := false
		for _, pl := range p.Platforms {
			if pl.Type == "feishu" || pl.Type == "lark" {
				hasFeishu = true
				break
			}
		}
		workDir, _ := p.Agent.Options["work_dir"].(string)
		projects = append(projects, ProjectInfo{
			Name:      p.Name,
			AgentType: p.Agent.Type,
			WorkDir:   workDir,
			HasFeishu: hasFeishu,
		})
	}
	return &ConfigSummary{
		ConfigPath: a.cfgPath,
		DataDir:    a.cfg.DataDir,
		Language:   string(a.cfg.Language),
		Projects:   projects,
	}, nil
}

// HasConfig 检查配置文件是否存在
func (a *App) HasConfig() bool {
	return a.cfgPath != "" && fileExists(a.cfgPath)
}

// ListRegisteredAgents 返回所有已注册的 agent 类型
func (a *App) ListRegisteredAgents() []string {
	return core.ListRegisteredAgents()
}

// ListRegisteredPlatforms 返回所有已注册的 platform 类型
func (a *App) ListRegisteredPlatforms() []string {
	return core.ListRegisteredPlatforms()
}

// resolveConfigPath 与 cmd/cc-connect/main.go 逻辑一致
func resolveConfigPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	if fileExists("config.toml") {
		return "config.toml"
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".cc-connect", "config.toml")
		if fileExists(p) {
			return p
		}
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
```

创建文件 `client/app.go`

注意：需要在文件顶部添加 `import "os"` 和 `import "path/filepath"`。

- [ ] **Step 2: 运行 go mod tidy 确认依赖**

Run: `cd /e/project/cc-connect-pro/client && go mod tidy`
Expected: 依赖解析成功，无报错

---

### Task 4: 创建 plugin 导入文件

**Files:**
- Create: `client/plugin_platform_feishu.go`
- Create: `client/plugin_agent_claudecode.go`
- Create: `client/plugin_agent_codex.go`

- [ ] **Step 1: 创建飞书平台插件导入**

```go
//go:build !no_feishu

package main

import _ "github.com/chenhg5/cc-connect/platform/feishu"
```

创建文件 `client/plugin_platform_feishu.go`

- [ ] **Step 2: 创建 claudecode agent 插件导入**

```go
//go:build !no_claudecode

package main

import _ "github.com/chenhg5/cc-connect/agent/claudecode"
```

创建文件 `client/plugin_agent_claudecode.go`

- [ ] **Step 3: 创建 codex agent 插件导入**

```go
//go:build !no_codex

package main

import _ "github.com/chenhg5/cc-connect/agent/codex"
```

创建文件 `client/plugin_agent_codex.go`

- [ ] **Step 4: 运行 go build 确认导入成功**

Run: `cd /e/project/cc-connect-pro/client && go build -tags '!no_feishu !no_claudecode !no_codex' .`
Expected: 编译成功（可能有 frontend/dist 不存在的错误，这是预期的，因为前端还没创建）

---

### Task 5: 创建 service.go — 服务状态机和 Engine 管理

**Files:**
- Create: `client/service.go`

- [ ] **Step 1: 创建 service.go**

```go
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/chenhg5/cc-connect/config"
	"github.com/chenhg5/cc-connect/core"
	"github.com/wailsapp/wails/v3/pkg/runtime"
)

// Service 管理 cc-connect 服务进程的生命周期
type Service struct {
	app     *App
	mu      sync.Mutex
	engines []*core.Engine
	status  ServiceStatus
	errMsg  string
	cancel  context.CancelFunc
}

// NewService 创建 Service 实例
func NewService(app *App) *Service {
	return &Service{
		app:    app,
		status: StatusIdle,
	}
}

// Start 启动所有 Engine
func (s *Service) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusRunning || s.status == StatusStarting {
		return fmt.Errorf("服务已在运行中")
	}

	s.status = StatusStarting
	s.errMsg = ""
	runtime.EventsEmit(s.app.ctx, "service:status", StatusStarting)

	cfgPath := s.app.cfgPath
	if cfgPath == "" {
		s.status = StatusError
		s.errMsg = "未找到配置文件"
		runtime.EventsEmit(s.app.ctx, "service:status", StatusError)
		runtime.EventsEmit(s.app.ctx, "service:error", s.errMsg)
		return fmt.Errorf(s.errMsg)
	}

	// 加载配置
	config.ConfigPath = cfgPath
	cfg, err := config.Load(cfgPath)
	if err != nil {
		s.status = StatusError
		s.errMsg = fmt.Sprintf("加载配置失败: %v", err)
		runtime.EventsEmit(s.app.ctx, "service:status", StatusError)
		runtime.EventsEmit(s.app.ctx, "service:error", s.errMsg)
		return err
	}
	s.app.cfg = cfg

	if len(cfg.Projects) == 0 {
		s.status = StatusError
		s.errMsg = "配置中没有项目"
		runtime.EventsEmit(s.app.ctx, "service:status", StatusError)
		runtime.EventsEmit(s.app.ctx, "service:error", s.errMsg)
		return fmt.Errorf(s.errMsg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// 为每个项目创建 Engine
	var engines []*core.Engine
	for _, proj := range cfg.Projects {
		engine, err := s.createEngine(ctx, cfg, proj)
		if err != nil {
			slog.Error("创建 Engine 失败", "project", proj.Name, "error", err)
			continue
		}
		engines = append(engines, engine)
	}

	if len(engines) == 0 {
		s.status = StatusError
		s.errMsg = "所有项目 Engine 创建失败"
		runtime.EventsEmit(s.app.ctx, "service:status", StatusError)
		runtime.EventsEmit(s.app.ctx, "service:error", s.errMsg)
		return fmt.Errorf(s.errMsg)
	}

	// 启动所有 Engine
	for _, e := range engines {
		if err := e.Start(); err != nil {
			slog.Error("启动 Engine 失败", "engine", e.Name(), "error", err)
		}
	}

	s.engines = engines
	s.status = StatusRunning
	runtime.EventsEmit(s.app.ctx, "service:status", StatusRunning)
	slog.Info("cc-connect 服务启动成功", "engines", len(engines))
	return nil
}

// Stop 停止所有 Engine
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status != StatusRunning {
		return fmt.Errorf("服务未在运行中")
	}

	s.status = StatusStopping
	runtime.EventsEmit(s.app.ctx, "service:status", StatusStopping)

	var errs []error
	for _, e := range s.engines {
		if err := e.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.engines = nil
	s.app.cfg = nil
	s.status = StatusIdle
	runtime.EventsEmit(s.app.ctx, "service:status", StatusIdle)

	if len(errs) > 0 {
		return fmt.Errorf("停止时发生错误: %v", errs)
	}
	slog.Info("cc-connect 服务已停止")
	return nil
}

// Restart 重启服务
func (s *Service) Restart() error {
	if s.status == StatusRunning {
		if err := s.Stop(); err != nil {
			return err
		}
	}
	return s.Start()
}

// GetEngines 返回当前所有 Engine
func (s *Service) GetEngines() []*core.Engine {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engines
}

// createEngine 为单个项目创建 Engine（复用 cmd/cc-connect/main.go 的核心逻辑）
func (s *Service) createEngine(ctx context.Context, cfg *config.Config, proj config.ProjectConfig) (*core.Engine, error) {
	// 构建 agent options
	agentOpts := make(map[string]any, len(proj.Agent.Options)+2)
	for k, v := range proj.Agent.Options {
		agentOpts[k] = v
	}
	agentOpts["cc_data_dir"] = cfg.DataDir
	agentOpts["cc_project"] = proj.Name

	// 创建 Agent
	agent, err := core.CreateAgent(proj.Agent.Type, agentOpts)
	if err != nil {
		return nil, fmt.Errorf("创建 Agent %s 失败: %w", proj.Agent.Type, err)
	}

	// 创建 Platforms
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
			slog.Warn("创建 Platform 失败", "type", pc.Type, "error", err)
			continue
		}
		platforms = append(platforms, p)
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("项目 %s 没有 Platform 可用", proj.Name)
	}

	// 确定 session 存储路径
	sessionFile := filepath.Join(cfg.DataDir, "sessions", proj.Name+".json")
	os.MkdirAll(filepath.Dir(sessionFile), 0o755)

	// 创建 Engine
	engine := core.NewEngine(proj.Name, agent, platforms, sessionFile, core.Language(cfg.Language))

	// 设置基础 Engine 配置
	engine.SetDataDir(cfg.DataDir)
	engine.SetAttachmentSendEnabled(cfg.AttachmentSend != "off")

	workDir, _ := proj.Agent.Options["work_dir"].(string)
	if workDir != "" {
		engine.SetBaseWorkDir(workDir)
	}

	// 设置显示配置（使用默认值，后续子项目会扩展）
	engine.SetDisplayConfig(core.DisplayCfg{
		Mode:             "full",
		ThinkingMessages: true,
		ToolMessages:     true,
	})

	return engine, nil
}
```

创建文件 `client/service.go`

- [ ] **Step 2: 在 app.go 中添加 StartService/StopService/RestartService 方法**

在 `client/app.go` 的 App 结构体中添加以下方法：

```go
// StartService 启动服务
func (a *App) StartService() error {
	return a.service.Start()
}

// StopService 停止服务
func (a *App) StopService() error {
	return a.service.Stop()
}

// RestartService 重启服务
func (a *App) RestartService() error {
	return a.service.Restart()
}
```

- [ ] **Step 3: 运行 go vet 检查**

Run: `cd /e/project/cc-connect-pro/client && go vet ./...`
Expected: 无报错

---

### Task 6: 创建 service_test.go — 状态机单元测试

**Files:**
- Create: `client/service_test.go`

- [ ] **Step 1: 创建测试文件**

```go
package main

import (
	"testing"
)

func TestServiceStatusTransitions(t *testing.T) {
	app := &App{}
	s := NewService(app)

	// 初始状态应该是 idle
	if s.status != StatusIdle {
		t.Errorf("期望初始状态为 idle，实际为 %s", s.status)
	}

	// idle → starting 应该失败（没有配置）
	err := s.Start()
	if err == nil {
		t.Error("期望无配置时 Start 返回错误")
	}
	if s.status != StatusError {
		t.Errorf("期望状态为 error，实际为 %s", s.status)
	}

	// error 状态下 Stop 应该失败
	err = s.Stop()
	if err == nil {
		t.Error("期望非运行状态下 Stop 返回错误")
	}
}

func TestServiceStatusConstants(t *testing.T) {
	statuses := map[ServiceStatus]string{
		StatusIdle:     "idle",
		StatusStarting: "starting",
		StatusRunning:  "running",
		StatusStopping: "stopping",
		StatusError:    "error",
	}
	for status, expected := range statuses {
		if string(status) != expected {
			t.Errorf("期望 %s，实际为 %s", expected, string(status))
		}
	}
}
```

创建文件 `client/service_test.go`

- [ ] **Step 2: 运行测试**

Run: `cd /e/project/cc-connect-pro/client && go test -run TestService -v`
Expected: 测试通过（注意：因为 Wails runtime 在测试中不可用，EventsEmit 调用需要处理）

- [ ] **Step 3: 修复测试中 Wails runtime 问题**

在 `service.go` 中，`runtime.EventsEmit` 在测试环境下会 panic（因为 ctx 为 nil）。添加空值检查：

在 service.go 的每个 `runtime.EventsEmit` 调用前添加保护：

```go
func emitEvent(ctx context.Context, name string, data ...any) {
	if ctx != nil {
		runtime.EventsEmit(ctx, name, data...)
	}
}
```

将 service.go 中所有 `runtime.EventsEmit(s.app.ctx, ...)` 替换为 `emitEvent(s.app.ctx, ...)`。

- [ ] **Step 4: 重新运行测试**

Run: `cd /e/project/cc-connect-pro/client && go test -run TestService -v`
Expected: 两个测试都 PASS

- [ ] **Step 5: Commit**

```bash
cd /e/project/cc-connect-pro
git add client/go.mod client/go.sum client/main.go client/app.go client/service.go client/service_test.go client/plugin_platform_feishu.go client/plugin_agent_claudecode.go client/plugin_agent_codex.go
git commit -m "feat(client): add Wails scaffolding with embedded service engine"
```

---

### Task 7: 创建 Wails 前端骨架

**Files:**
- Create: `client/frontend/package.json`
- Create: `client/frontend/vite.config.ts`
- Create: `client/frontend/tsconfig.json`
- Create: `client/frontend/index.html`
- Create: `client/frontend/postcss.config.js`
- Create: `client/frontend/tailwind.config.ts`
- Create: `client/frontend/src/main.tsx`
- Create: `client/frontend/src/styles/index.css`

- [ ] **Step 1: 创建 package.json**

```json
{
  "name": "cc-connect-client",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^19.0.0",
    "react-dom": "^19.0.0",
    "react-router-dom": "^7.1.0",
    "@wails/runtime": "latest",
    "lucide-react": "^0.460.0"
  },
  "devDependencies": {
    "@types/react": "^19.0.0",
    "@types/react-dom": "^19.0.0",
    "@vitejs/plugin-react": "^4.3.0",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.4.49",
    "tailwindcss": "^3.4.17",
    "typescript": "^5.7.0",
    "vite": "^6.0.0"
  }
}
```

创建文件 `client/frontend/package.json`

- [ ] **Step 2: 创建 vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  server: {
    port: 5173,
    strictPort: true,
  },
  envPrefix: ['VITE_', 'WAILS_'],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
})
```

创建文件 `client/frontend/vite.config.ts`

- [ ] **Step 3: 创建 tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src"]
}
```

创建文件 `client/frontend/tsconfig.json`

- [ ] **Step 4: 创建 index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>cc-connect</title>
  </head>
  <body class="bg-canvas text-primary">
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

创建文件 `client/frontend/index.html`

- [ ] **Step 5: 创建 PostCSS 和 Tailwind 配置**

```javascript
// postcss.config.js
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
```

创建文件 `client/frontend/postcss.config.js`

```typescript
// tailwind.config.ts — 视觉设计系统 token
import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        canvas: {
          DEFAULT: '#FFFFFF',
          dark: '#1C1C1E',
        },
        surface: {
          DEFAULT: '#F5F5F7',
          dark: '#2C2C2E',
        },
        accent: {
          DEFAULT: '#5856D6',
          dark: '#7B79E0',
        },
        success: {
          DEFAULT: '#34C759',
          dark: '#30D158',
        },
        warning: {
          DEFAULT: '#FF9500',
          dark: '#FF9F0A',
        },
        primary: {
          DEFAULT: '#1D1D1F',
          dark: '#F5F5F7',
        },
        secondary: {
          DEFAULT: '#86868B',
          dark: '#A1A1A6',
        },
      },
      fontFamily: {
        sans: ['-apple-system', 'BlinkMacSystemFont', '"Segoe UI"', 'Roboto', 'sans-serif'],
        mono: ['ui-monospace', '"SF Mono"', '"Cascadia Code"', 'Menlo', 'monospace'],
      },
      fontSize: {
        'large-title': ['28px', { lineHeight: '34px' }],
        'title': ['20px', { lineHeight: '25px' }],
        'headline': ['17px', { lineHeight: '22px', fontWeight: '600' }],
        'body': ['15px', { lineHeight: '20px' }],
        'caption': ['12px', { lineHeight: '16px' }],
        'mini': ['10px', { lineHeight: '13px' }],
      },
      borderRadius: {
        DEFAULT: '8px',
        sm: '6px',
      },
      spacing: {
        sidebar: '220px',
        content: '32px',
        card: '16px',
      },
    },
  },
  plugins: [],
} satisfies Config
```

创建文件 `client/frontend/tailwind.config.ts`

- [ ] **Step 6: 创建 CSS 入口**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  /* 苹果风格的滚动条 */
  ::-webkit-scrollbar {
    width: 6px;
  }
  ::-webkit-scrollbar-track {
    background: transparent;
  }
  ::-webkit-scrollbar-thumb {
    background: #86868B;
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: #5856D6;
  }
}
```

创建文件 `client/frontend/src/styles/index.css`

- [ ] **Step 7: 创建 React 入口**

```tsx
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './styles/index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
```

创建文件 `client/frontend/src/main.tsx`

- [ ] **Step 8: 安装前端依赖**

Run: `cd /e/project/cc-connect-pro/client/frontend && npm install`
Expected: 依赖安装成功

---

### Task 8: 创建前端类型定义和 hooks

**Files:**
- Create: `client/frontend/src/lib/types.ts`
- Create: `client/frontend/src/hooks/useServiceStatus.ts`

- [ ] **Step 1: 创建类型定义**

```typescript
// types.ts — 前端共享类型

export type ServiceStatus = 'idle' | 'starting' | 'running' | 'stopping' | 'error'

export interface ConfigSummary {
  configPath: string
  dataDir: string
  language: string
  projects: ProjectInfo[]
}

export interface ProjectInfo {
  name: string
  agentType: string
  workDir: string
  hasFeishu: boolean
}

export interface ServiceError {
  message: string
  timestamp: number
}
```

创建文件 `client/frontend/src/lib/types.ts`

- [ ] **Step 2: 创建 useServiceStatus hook**

```typescript
import { useState, useEffect } from 'react'
import { EventsOn } from '@wails/runtime'
import type { ServiceStatus } from '../lib/types'

export function useServiceStatus(initial: ServiceStatus = 'idle') {
  const [status, setStatus] = useState<ServiceStatus>(initial)

  useEffect(() => {
    const unsubscribe = EventsOn('service:status', (newStatus: ServiceStatus) => {
      setStatus(newStatus)
    })
    return () => {
      if (typeof unsubscribe === 'function') {
        unsubscribe()
      }
    }
  }, [])

  return status
}
```

创建文件 `client/frontend/src/hooks/useServiceStatus.ts`

---

### Task 9: 创建基础 UI 组件

**Files:**
- Create: `client/frontend/src/components/StatusDot.tsx`
- Create: `client/frontend/src/components/SegmentedControl.tsx`
- Create: `client/frontend/src/components/ConnectionBridge.tsx`

- [ ] **Step 1: 创建 StatusDot 组件**

```tsx
interface StatusDotProps {
  status: 'active' | 'idle' | 'error' | 'warning'
  label?: string
  size?: number
}

const colorMap = {
  active: 'bg-success',
  idle: 'bg-secondary',
  error: 'bg-red-500',
  warning: 'bg-warning',
}

export function StatusDot({ status, label, size = 6 }: StatusDotProps) {
  return (
    <span className="inline-flex items-center gap-2">
      <span
        className={`${colorMap[status]} rounded-full`}
        style={{ width: size, height: size }}
      />
      {label && <span className="text-caption text-secondary">{label}</span>}
    </span>
  )
}
```

创建文件 `client/frontend/src/components/StatusDot.tsx`

- [ ] **Step 2: 创建 SegmentedControl 组件**

```tsx
interface SegmentedControlProps<T extends string> {
  options: { label: string; value: T }[]
  value: T
  onChange: (value: T) => void
}

export function SegmentedControl<T extends string>({
  options,
  value,
  onChange,
}: SegmentedControlProps<T>) {
  return (
    <div className="inline-flex rounded-sm bg-surface p-1">
      {options.map((opt) => (
        <button
          key={opt.value}
          className={`px-4 py-1.5 rounded-sm text-body transition-colors ${
            opt.value === value
              ? 'bg-accent text-white'
              : 'text-secondary hover:text-primary'
          }`}
          onClick={() => onChange(opt.value)}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
```

创建文件 `client/frontend/src/components/SegmentedControl.tsx`

- [ ] **Step 3: 创建 ConnectionBridge 签名动画组件**

```tsx
import { useEffect, useRef } from 'react'
import type { ServiceStatus } from '../lib/types'

interface ConnectionBridgeProps {
  status: ServiceStatus
}

export function ConnectionBridge({ status }: ConnectionBridgeProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const animRef = useRef<number>(0)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')!
    const w = canvas.width = canvas.offsetWidth * 2
    const h = canvas.height = canvas.offsetHeight * 2
    ctx.scale(2, 2)

    const lineY = canvas.offsetHeight / 2
    const startX = 60
    const endX = canvas.offsetWidth - 60
    const dotPos = { x: startX }

    function draw() {
      ctx.clearRect(0, 0, canvas.offsetWidth, canvas.offsetHeight)

      // 连接线
      const lineColor = status === 'running' ? '#5856D6' : '#86868B'
      ctx.strokeStyle = lineColor
      ctx.lineWidth = 1.5
      ctx.beginPath()
      ctx.moveTo(startX, lineY)
      ctx.lineTo(endX, lineY)
      ctx.stroke()

      // 节点
      ctx.fillStyle = status === 'running' ? '#5856D6' : '#86868B'
      ctx.beginPath()
      ctx.arc(startX, lineY, 8, 0, Math.PI * 2)
      ctx.fill()
      ctx.beginPath()
      ctx.arc(endX, lineY, 8, 0, Math.PI * 2)
      ctx.fill()

      // Agent 标签
      ctx.font = '12px -apple-system, sans-serif'
      ctx.fillStyle = status === 'running' ? '#1D1D1F' : '#86868B'
      ctx.textAlign = 'center'
      ctx.fillText('Agent', startX, lineY - 20)
      ctx.fillText('飞书', endX, lineY - 20)

      // 流动光点（仅运行时）
      if (status === 'running') {
        dotPos.x += 0.5
        if (dotPos.x > endX) dotPos.x = startX

        ctx.fillStyle = '#5856D6'
        ctx.beginPath()
        ctx.arc(dotPos.x, lineY, 3, 0, Math.PI * 2)
        ctx.fill()

        // 光点拖尾
        ctx.fillStyle = 'rgba(88, 86, 214, 0.3)'
        ctx.beginPath()
        ctx.arc(dotPos.x - 6, lineY, 2, 0, Math.PI * 2)
        ctx.fill()
      }

      animRef.current = requestAnimationFrame(draw)
    }

    draw()

    return () => {
      cancelAnimationFrame(animRef.current)
    }
  }, [status])

  return (
    <canvas
      ref={canvasRef}
      className="w-full h-20"
    />
  )
}
```

创建文件 `client/frontend/src/components/ConnectionBridge.tsx`

---

### Task 10: 创建侧边栏布局组件

**Files:**
- Create: `client/frontend/src/layouts/SidebarLayout.tsx`

- [ ] **Step 1: 创建 SidebarLayout**

```tsx
import { NavLink } from 'react-router-dom'
import { Activity, MessageSquare, Bot, Monitor } from 'lucide-react'
import type { ServiceStatus } from '../lib/types'
import { StatusDot } from '../components/StatusDot'

interface SidebarLayoutProps {
  status: ServiceStatus
  children: React.ReactNode
}

const navItems = [
  { path: '/', label: '总览', icon: Activity },
  { path: '/feishu', label: '飞书配置', icon: MessageSquare },
  { path: '/agents', label: 'Agent 管理', icon: Bot },
  { path: '/monitor', label: '监控', icon: Monitor },
]

export function SidebarLayout({ status, children }: SidebarLayoutProps) {
  return (
    <div className="flex h-screen bg-canvas">
      {/* 侧边栏 */}
      <aside className="w-sidebar bg-surface flex flex-col">
        {/* 品牌 + 全局状态 */}
        <div className="px-5 pt-6 pb-4">
          <div className="text-large-title font-semibold text-primary">cc-connect</div>
          <div className="mt-2">
            <StatusDot
              status={status === 'running' ? 'active' : status === 'error' ? 'error' : 'idle'}
              label={status === 'running' ? '运行中' : status === 'error' ? '错误' : '已停止'}
            />
          </div>
        </div>

        {/* 导航 */}
        <nav className="flex-1 px-2">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-sm my-0.5 text-body transition-colors ${
                  isActive
                    ? 'bg-surface text-primary relative before:absolute before:left-0 before:top-1 before:bottom-1 before:w-[3px] before:bg-accent before:rounded-full'
                    : 'text-secondary hover:text-primary'
                }`
              }
            >
              <item.icon size={18} />
              {item.label}
            </NavLink>
          ))}
        </nav>

        {/* 底部版本信息 */}
        <div className="px-5 py-4 text-mini text-secondary">
          v0.1.0
        </div>
      </aside>

      {/* 内容区 */}
      <main className="flex-1 p-content overflow-auto">
        {children}
      </main>
    </div>
  )
}
```

创建文件 `client/frontend/src/layouts/SidebarLayout.tsx`

---

### Task 11: 创建 Dashboard 页面

**Files:**
- Create: `client/frontend/src/pages/Dashboard.tsx`

- [ ] **Step 1: 创建 Dashboard 页面**

```tsx
import { StartService, StopService, RestartService, GetServiceStatus } from '../wailsjs/go/main/App'
import { ConnectionBridge } from '../components/ConnectionBridge'
import { StatusDot } from '../components/StatusDot'
import { useServiceStatus } from '../hooks/useServiceStatus'

export function Dashboard() {
  const status = useServiceStatus()

  const handleStart = async () => {
    try {
      await StartService()
    } catch (err) {
      console.error('启动失败:', err)
    }
  }

  const handleStop = async () => {
    try {
      await StopService()
    } catch (err) {
      console.error('停止失败:', err)
    }
  }

  const handleRestart = async () => {
    try {
      await RestartService()
    } catch (err) {
      console.error('重启失败:', err)
    }
  }

  return (
    <div className="space-y-card">
      {/* 签名动画 */}
      <ConnectionBridge status={status} />

      {/* 服务状态卡片 */}
      <div className="bg-surface rounded-DEFAULT p-6">
        <h2 className="text-title text-primary mb-4">服务状态</h2>
        <div className="flex items-center gap-6 mb-6">
          <StatusDot
            status={status === 'running' ? 'active' : status === 'error' ? 'error' : 'idle'}
            label={status === 'running' ? '运行中' : status === 'error' ? '错误' : status === 'starting' ? '启动中' : '已停止'}
            size={8}
          />
          <span className="font-mono text-body text-secondary">
            {status}
          </span>
        </div>

        {/* 操作按钮 */}
        <div className="flex gap-3">
          {status === 'idle' || status === 'error' ? (
            <button
              onClick={handleStart}
              className="px-5 py-2 rounded-sm bg-accent text-white text-body hover:bg-accent/90 transition-colors"
            >
              启动服务
            </button>
          ) : status === 'running' ? (
            <>
              <button
                onClick={handleStop}
                className="px-5 py-2 rounded-sm bg-surface text-secondary text-body hover:text-primary transition-colors"
              >
                停止服务
              </button>
              <button
                onClick={handleRestart}
                className="px-5 py-2 rounded-sm bg-surface text-secondary text-body hover:text-primary transition-colors"
              >
                重启服务
              </button>
            </>
          ) : null}
        </div>
      </div>

      {/* 占位：后续子项目会填充 Platform 状态、Agent 状态等 */}
      <div className="bg-surface rounded-DEFAULT p-6">
        <h2 className="text-title text-primary mb-4">平台连接</h2>
        <p className="text-body text-secondary">服务启动后显示各平台连接状态</p>
      </div>

      <div className="bg-surface rounded-DEFAULT p-6">
        <h2 className="text-title text-primary mb-4">Agent 状态</h2>
        <p className="text-body text-secondary">服务启动后显示各 Agent 活跃信息</p>
      </div>
    </div>
  )
}
```

创建文件 `client/frontend/src/pages/Dashboard.tsx`

---

### Task 12: 创建 Welcome 引导页

**Files:**
- Create: `client/frontend/src/pages/Welcome.tsx`

- [ ] **Step 1: 创建 Welcome 引导页**

```tsx
import { useState } from 'react'
import { HasConfig, ListRegisteredAgents, StartService } from '../wailsjs/go/main/App'

export function Welcome() {
  const [appId, setAppId] = useState('')
  const [appSecret, setAppSecret] = useState('')
  const [agentType, setAgentType] = useState('')
  const [projectName, setProjectName] = useState('my-project')
  const [workDir, setWorkDir] = useState('')
  const [agents, setAgents] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // 加载可用 agent 列表
  useEffect(() => {
    ListRegisteredAgents().then((list) => {
      setAgents(list)
      if (list.length > 0 && !agentType) {
        setAgentType(list[0])
      }
    })
  }, [])

  const handleSetup = async () => {
    if (!appId || !appSecret) {
      setError('请填写飞书 App ID 和 App Secret')
      return
    }
    if (!projectName) {
      setError('请填写项目名称')
      return
    }

    setLoading(true)
    setError('')

    try {
      // 调用 Go 侧的 CreateProjectAndStart 方法（下一步在 app.go 中添加）
      await CreateProjectWithFeishu({
        projectName,
        agentType,
        appId,
        appSecret,
        workDir,
      })
      await StartService()
    } catch (err) {
      setError(String(err))
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-canvas">
      <div className="w-full max-w-md p-content">
        <h1 className="text-large-title text-primary mb-2">欢迎使用 cc-connect</h1>
        <p className="text-body text-secondary mb-8">
          配置飞书应用凭证，即可开始使用
        </p>

        {/* 飞书凭证 */}
        <div className="space-y-4">
          <div>
            <label className="text-caption text-secondary block mb-1.5">App ID</label>
            <input
              type="text"
              value={appId}
              onChange={(e) => setAppId(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-secondary/30 text-body text-primary focus:border-accent focus:outline-none transition-colors"
              placeholder="cli_xxxxxxxxxxxx"
            />
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1.5">App Secret</label>
            <input
              type="password"
              value={appSecret}
              onChange={(e) => setAppSecret(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-secondary/30 text-body text-primary focus:border-accent focus:outline-none transition-colors"
              placeholder="飞书应用密钥"
            />
          </div>

          {/* 项目配置 */}
          <div>
            <label className="text-caption text-secondary block mb-1.5">项目名称</label>
            <input
              type="text"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-secondary/30 text-body text-primary focus:border-accent focus:outline-none transition-colors"
            />
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1.5">Agent 类型</label>
            <select
              value={agentType}
              onChange={(e) => setAgentType(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-secondary/30 text-body text-primary focus:border-accent focus:outline-none transition-colors bg-canvas"
            >
              {agents.map((a) => (
                <option key={a} value={a}>{a}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1.5">工作目录（可选）</label>
            <input
              type="text"
              value={workDir}
              onChange={(e) => setWorkDir(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-secondary/30 text-body text-primary focus:border-accent focus:outline-none transition-colors"
              placeholder="/path/to/project"
            />
          </div>
        </div>

        {/* 错误信息 */}
        {error && (
          <p className="mt-4 text-body text-warning">{error}</p>
        )}

        {/* 启动按钮 */}
        <button
          onClick={handleSetup}
          disabled={loading}
          className="mt-6 w-full px-5 py-2.5 rounded-sm bg-accent text-white text-body hover:bg-accent/90 transition-colors disabled:opacity-50"
        >
          {loading ? '正在配置...' : '创建并启动'}
        </button>
      </div>
    </div>
  )
}
```

创建文件 `client/frontend/src/pages/Welcome.tsx`

注意：此文件需要 `import { useEffect } from 'react'` 和一个 `CreateProjectWithFeishu` 的 Wails 绑定调用，这在 Task 13 中添加。

---

### Task 13: 在 app.go 中添加 CreateProjectWithFeishu 方法

**Files:**
- Modify: `client/app.go`

- [ ] **Step 1: 在 app.go 中添加 CreateProjectWithFeishu 方法和相关类型**

```go
// CreateProjectWithFeishuOpts 创建项目+飞书平台的选项
type CreateProjectWithFeishuOpts struct {
	ProjectName string `json:"projectName"`
	AgentType   string `json:"agentType"`
	AppID       string `json:"appId"`
	AppSecret   string `json:"appSecret"`
	WorkDir     string `json:"workDir"`
}

// CreateProjectWithFeishu 创建项目并配置飞书凭证，然后写入 config.toml
func (a *App) CreateProjectWithFeishu(opts CreateProjectWithFeishuOpts) error {
	if opts.ProjectName == "" {
		return fmt.Errorf("项目名称不能为空")
	}
	if opts.AppID == "" || opts.AppSecret == "" {
		return fmt.Errorf("飞书 App ID 和 App Secret 不能为空")
	}

	// 确保数据目录存在
	dataDir := filepath.Join(homeDir(), ".cc-connect")
	os.MkdirAll(dataDir, 0o755)

	// 确定配置文件路径
	cfgPath := filepath.Join(dataDir, "config.toml")

	// 如果配置文件不存在，创建空配置
	if !fileExists(cfgPath) {
		if err := os.WriteFile(cfgPath, []byte(""), 0o644); err != nil {
			return fmt.Errorf("创建配置文件失败: %w", err)
		}
	}

	// 设置全局 ConfigPath（config 包的 Save 函数依赖它）
	config.ConfigPath = cfgPath
	a.cfgPath = cfgPath

	// 创建项目+飞书平台
	result, err := config.EnsureProjectWithFeishuPlatform(config.EnsureProjectWithFeishuOptions{
		ProjectName: opts.ProjectName,
		PlatformType: "feishu",
		AgentType:    opts.AgentType,
		WorkDir:      opts.WorkDir,
	})
	if err != nil {
		return fmt.Errorf("创建项目失败: %w", err)
	}

	// 写入飞书凭证
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

	slog.Info("项目创建成功", "project", result.ProjectName, "created", result.Created)
	return nil
}

func homeDir() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}
```

添加到 `client/app.go` 文件中。

- [ ] **Step 2: 运行 go vet 检查**

Run: `cd /e/project/cc-connect-pro/client && go vet ./...`
Expected: 无报错

---

### Task 14: 创建 App.tsx — 主布局和路由

**Files:**
- Create: `client/frontend/src/App.tsx`

- [ ] **Step 1: 创建 App.tsx**

```tsx
import { HashRouter, Routes, Route } from 'react-router-dom'
import { SidebarLayout } from './layouts/SidebarLayout'
import { Dashboard } from './pages/Dashboard'
import { Welcome } from './pages/Welcome'
import { useServiceStatus } from './hooks/useServiceStatus'
import { HasConfig } from './wailsjs/go/main/App'
import { useState, useEffect } from 'react'

export default function App() {
  const [hasConfig, setHasConfig] = useState<boolean | null>(null)
  const status = useServiceStatus()

  useEffect(() => {
    HasConfig().then((result) => {
      setHasConfig(result)
    })
  }, [])

  // 还在检查配置时
  if (hasConfig === null) {
    return (
      <div className="flex items-center justify-center h-screen bg-canvas">
        <p className="text-body text-secondary">正在检查配置...</p>
      </div>
    )
  }

  // 没有配置 → 显示欢迎引导页
  if (!hasConfig) {
    return <Welcome onConfigCreated={() => setHasConfig(true)} />
  }

  // 有配置 → 显示主界面
  return (
    <HashRouter>
      <SidebarLayout status={status}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/feishu" element={<FeishuPlaceholder />} />
          <Route path="/agents" element={<AgentsPlaceholder />} />
          <Route path="/monitor" element={<MonitorPlaceholder />} />
        </Routes>
      </SidebarLayout>
    </HashRouter>
  )
}

// 占位页面——后续子项目会替换为完整实现
function FeishuPlaceholder() {
  return (
    <div className="bg-surface rounded-DEFAULT p-6">
      <h2 className="text-title text-primary">飞书配置</h2>
      <p className="text-body text-secondary mt-2">子项目 3 将实现完整的飞书凭证与工作区管理页面</p>
    </div>
  )
}

function AgentsPlaceholder() {
  return (
    <div className="bg-surface rounded-DEFAULT p-6">
      <h2 className="text-title text-primary">Agent 管理</h2>
      <p className="text-body text-secondary mt-2">子项目 4 将实现完整的 Agent 状态与任务队列页面</p>
    </div>
  )
}

function MonitorPlaceholder() {
  return (
    <div className="bg-surface rounded-DEFAULT p-6">
      <h2 className="text-title text-primary">监控</h2>
      <p className="text-body text-secondary mt-2">子项目 5 将实现完整的健康监控页面</p>
    </div>
  )
}
```

创建文件 `client/frontend/src/App.tsx`

注意：Welcome 组件需要 `onConfigCreated` prop，需要修改 Welcome.tsx 添加此 prop。

- [ ] **Step 2: 更新 Welcome.tsx 接收 onConfigCreated prop**

修改 `client/frontend/src/pages/Welcome.tsx`：

在 Welcome 函数签名中添加 prop：
```tsx
interface WelcomeProps {
  onConfigCreated: () => void
}

export function Welcome({ onConfigCreated }: WelcomeProps) {
```

在 `handleSetup` 成功后调用：
```tsx
    try {
      await CreateProjectWithFeishu({ ... })
      await StartService()
      onConfigCreated()  // 通知 App 配置已创建
    } catch (err) {
```

---

### Task 15: 创建 wails.json 配置

**Files:**
- Create: `client/wails.json`

- [ ] **Step 1: 创建 wails.json**

```json
{
  "$schema": "https://wails.io/schemas/wails.v3.json",
  "name": "cc-connect",
  "version": "0.1.0",
  "author": {
    "name": "cc-connect"
  },
  "description": "cc-connect Desktop Client",
  "frontend": {
    "dir": "frontend",
    "install": "npm install",
    "build": "npm run build",
    "dev": "npm run dev",
    "bin": "frontend/dist"
  },
  "build": {
    "outputfilename": "cc-connect",
    "obfuscated": false,
    "ldflags": [],
    "tags": [],
    "strip": false
  },
  "bindings": {
    "dir": "bindings"
  }
}
```

创建文件 `client/wails.json`

---

### Task 16: 创建 Makefile

**Files:**
- Create: `client/Makefile`

- [ ] **Step 1: 创建 Makefile**

```makefile
.PHONY: dev build clean test

dev:
	wails3 dev

build:
	cd frontend && npm run build
	wails3 build

build-windows:
	cd frontend && npm run build
	wails3 build -platform windows/amd64

build-darwin:
	cd frontend && npm run build
	wails3 build -platform darwin/universal

clean:
	rm -rf frontend/dist frontend/node_modules build

test:
	go test ./...

frontend-install:
	cd frontend && npm install
```

创建文件 `client/Makefile`

---

### Task 17: 首次构建验证

- [ ] **Step 1: 构建前端**

Run: `cd /e/project/cc-connect-pro/client/frontend && npm run build`
Expected: Vite 构建成功，dist 目录生成

- [ ] **Step 2: 生成 Wails 绑定**

Run: `cd /e/project/cc-connect-pro/client && wails3 generate bindings`
Expected: `frontend/wailsjs/go/main/` 目录生成，包含 App.js 和 App.d.ts

- [ ] **Step 3: 运行 go build 确认编译成功**

Run: `cd /e/project/cc-connect-pro/client && go build .`
Expected: 编译成功，生成可执行文件

- [ ] **Step 4: 运行 Go 测试**

Run: `cd /e/project/cc-connect-pro/client && go test ./...`
Expected: 所有测试通过

- [ ] **Step 5: 确认父项目测试仍然通过**

Run: `cd /e/project/cc-connect-pro && go test ./core/ ./config/ -v`
Expected: 现有测试全部通过，不受 client/ 影响

---

### Task 18: 最终 Commit

- [ ] **Step 1: 添加所有新文件**

```bash
cd /e/project/cc-connect-pro
git add client/
git commit -m "feat(client): add Wails desktop client scaffold with embedded service engine

- Wails v3 project in client/ directory with own go.mod (replace directive to parent)
- service.go: service state machine (idle/starting/running/stopping/error)
- app.go: Go methods bound to frontend (service control, config, agent listing)
- Plugin imports for feishu, claudecode, codex agents/platforms
- React + TypeScript + Tailwind frontend with Apple HIG design tokens
- SidebarLayout with accent capsule navigation indicator
- Dashboard page with ConnectionBridge signature animation
- Welcome wizard for first-time Feishu credential setup
- StatusDot, SegmentedControl reusable components
- Unit tests for service state machine"
```

---

## 自审清单

**1. Spec 覆盖度：**
- Wails 项目搭建 ✓ Task 1-2
- 内嵌 Engine 启动/停止 ✓ Task 5-6 (service.go)
- 侧边栏导航 ✓ Task 10 (SidebarLayout)
- Dashboard 空壳 ✓ Task 11
- 视觉设计系统 ✓ Task 7 (tailwind.config.ts)
- 首次启动引导 ✓ Task 12-13 (Welcome + CreateProjectWithFeishu)
- 系统托盘 → 子项目 2，不在本计划范围

**2. Placeholder 扫描：** 无 TBD/TODO，所有步骤有具体代码 ✓

**3. 类型一致性：**
- `ServiceStatus` Go 侧 (`app.go`) 和 TS 侧 (`types.ts`) 一致 ✓
- `ConfigSummary`/`ProjectInfo` Go 和 TS 结构一致 ✓
- `CreateProjectWithFeishuOpts` Go 和 TS 调用参数名一致 ✓
