# 子项目 2：系统托盘 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 为 cc-connect 桌面客户端添加系统托盘功能——托盘图标、菜单、状态动态更新、关闭窗口隐藏到托盘而非退出。

**Architecture:** 使用 Wails v3 原生 `app.SystemTray.New()` API，在 `client/tray.go` 中封装托盘创建和菜单构建逻辑。修改 `main.go` 整合托盘设置。在 `app.go` 中添加窗口隐藏/显示和退出方法。

**Tech Stack:** Wails v3 SystemTray API, Go, PNG icon

---

### Task 1: 创建 tray.go — 托盘设置和菜单构建

**Files:**
- Create: `client/tray.go`

- [ ] **Step 1: 创建 tray.go**

```go
package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// setupTray 创建系统托盘并配置菜单
func setupTray(app *application.App, window *application.WebviewWindow, serviceApp *App) *application.SystemTray {
	tray := app.SystemTray.New()
	tray.SetTooltip("cc-connect")
	tray.SetIcon(defaultIconBytes)

	menu := application.NewMenu()

	// 状态显示（动态更新）
	statusItem := menu.Add("cc-connect - 已停止")
	statusItem.SetDisabled(true)

	menu.AddSeparator()

	// 显示窗口
	menu.Add("显示窗口").OnClick(func(ctx *application.Context) {
		window.Show()
		window.SetFocus()
	})

	// 重启服务
	menu.Add("重启服务").OnClick(func(ctx *application.Context) {
		serviceApp.RestartService()
	})

	menu.AddSeparator()

	// 退出
	menu.Add("退出").OnClick(func(ctx *application.Context) {
		serviceApp.Shutdown()
		app.Quit()
	})

	tray.SetMenu(menu)

	// 点击托盘图标显示窗口
	tray.OnClick(func() {
		window.Show()
		window.SetFocus()
	})

	return tray
}

// updateTrayStatus 更新托盘状态文字和图标
func updateTrayStatus(tray *application.SystemTray, menu *application.Menu, status ServiceStatus) {
	statusLabels := map[ServiceStatus]string{
		StatusIdle:     "cc-connect - 已停止",
		StatusStarting: "cc-connect - 启动中...",
		StatusRunning:  "cc-connect - 运行中",
		StatusStopping: "cc-connect - 停止中...",
		StatusError:    "cc-connect - 错误",
	}

	label := statusLabels[status]
	// 更新第一个菜单项的文字（状态项）
	if len(menu.Items) > 0 {
		menu.Items[0].SetLabel(label)
	}

	// 根据状态切换图标颜色
	if status == StatusRunning {
		tray.SetIcon(accentIconBytes)
	} else {
		tray.SetIcon(defaultIconBytes)
	}
}
```

- [ ] **Step 2: 创建图标文件和嵌入**

在 `client/icons.go` 中嵌入托盘图标 PNG：

```go
package main

import "embed"

//go:embed icons/tray-default.png
var defaultIconBytes []byte

//go:embed icons/tray-accent.png
var accentIconBytes []byte
```

需要创建两个 PNG 图标文件：
- `client/icons/tray-default.png` — 灰色连接符号（16x16 或 32x32）
- `client/icons/tray-accent.png` — Accent 色 (#5856D6) 连接符号

由于无法生成真实 PNG，使用简单的代码生成方案：在 tray.go 中用 Go 代码生成最小的 PNG 字节。

修改 `icons.go` 改为程序化生成图标：

```go
package main

// 生成最小 PNG 图标字节
// PNG 格式: signature + IHDR + IDAT + IEND

func generateDefaultIcon() []byte {
	// 生成 16x16 灰色圆点 PNG
	return generateCirclePNG(16, 16, 128, 128, 128, 255) // 灰色 RGBA
}

func generateAccentIcon() []byte {
	// 生成 16x16 accent 色圆点 PNG
	return generateCirclePNG(16, 16, 88, 86, 214, 255) // #5856D6 RGBA
}

var defaultIconBytes = generateDefaultIcon()
var accentIconBytes = generateAccentIcon()
```

实际上，用纯 Go 生成 PNG 比较复杂。更简单的方案：用 `image` 和 `image/png` 包生成，但这些依赖会增加复杂度。

最简方案：直接用 Go 的 `image` + `image/png` + `image/color` 标准库生成：

```go
package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

var defaultIconBytes = mustEncodePNG(generateIcon(color.RGBA{R: 134, G: 134, B: 139, A: 255}))
var accentIconBytes = mustEncodePNG(generateIcon(color.RGBA{R: 88, G: 86, B: 214, A: 255}))

func generateIcon(c color.Color) *image.RGBA {
	const size = 16
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// 画一个简单的连接符号：两个小圆+一条线
	// 背景透明
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{A: 0})
		}
	}
	// 左节点 (3, 8)
	drawCircle(img, 3, 8, 2, c)
	// 右节点 (12, 8)
	drawCircle(img, 12, 8, 2, c)
	// 连接线
	for x := 5; x <= 10; x++ {
		img.Set(x, 8, c)
		img.Set(x, 7, c)
	}
	return img
}

func drawCircle(img *image.RGBA, cx, cy, r int, c color.Color) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				img.Set(cx+dx, cy+dy, c)
			}
		}
	}
}

func mustEncodePNG(img image.Image) []byte {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}
```

---

### Task 2: 修改 main.go — 整合托盘和窗口管理

**Files:**
- Modify: `client/main.go`

- [ ] **Step 1: 修改 main.go**

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
			Handler: application.AssetFileServerFS(assets),
		},
	})

	serviceApp := &App{}
	app.RegisterService(application.NewService(serviceApp))

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "cc-connect",
		Width:  960,
		Height: 640,
		URL:    "/",
	})

	// 设置系统托盘
	tray := setupTray(app, window, serviceApp)
	serviceApp.tray = tray
	serviceApp.trayMenu = tray.Menu()  // 需要确认 Menu() 方法是否可用
	serviceApp.window = window

	// 关闭窗口时隐藏而非退出
	window.OnWindowClosing(func(window *application.WebviewWindow) bool {
		window.Hide()
		return false  // 返回 false 阻止关闭
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
```

注意：Wails v3 中 `OnWindowClosing` 可能使用不同的 API。需要确认实际的回调签名。如果 `OnWindowClosing` 不存在，可能需要用 `window.Events().On(application.WindowClosingEvent)` 或类似方式。

---

### Task 3: 修改 app.go — 添加托盘和窗口相关字段和方法

**Files:**
- Modify: `client/app.go`

- [ ] **Step 1: 在 App struct 中添加托盘和窗口字段**

```go
type App struct {
	ctx     context.Context
	service *Service
	status  ServiceStatus
	cfg     *config.Config
	cfgPath string
	tray    *application.SystemTray
	trayMenu *application.Menu
	window  *application.WebviewWindow
}
```

- [ ] **Step 2: 添加 Shutdown 方法**

```go
// Shutdown 优雅关闭所有 Engine 并准备退出
func (a *App) Shutdown() error {
	if a.service != nil {
		return a.service.Stop()
	}
	return nil
}
```

- [ ] **Step 3: 在 ServiceStartup 中监听状态变化并更新托盘**

在 `ServiceStartup` 中添加事件监听器，当服务状态变化时更新托盘图标和菜单文字：

```go
func (a *App) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	a.ctx = ctx
	a.status = StatusIdle
	a.service = NewService(a)

	// 监听服务状态事件，更新托盘
	application.Get().Event.On("service:status", func(data any) {
		if status, ok := data.(ServiceStatus); ok {
			a.status = status
			if a.tray != nil && a.trayMenu != nil {
				updateTrayStatus(a.tray, a.trayMenu, status)
			}
		}
	})

	return nil
}
```

注意：Wails v3 的 `Event.On` 方法签名可能不同。需要确认实际 API。可能是：
- `application.Get().Event.On(name, callback)`
- 或需要通过 `application.EventsOn(ctx, name, callback)`

---

### Task 4: 测试和验证

- [ ] **Step 1: 运行 go vet**

Run: `export PATH="/e/Go/bin:$HOME/go/bin:$PATH" && export GOPROXY=https://goproxy.cn,direct && cd /e/project/cc-connect-pro/client && go vet ./...`
Expected: 无报错

- [ ] **Step 2: 运行 go mod tidy**

Run: `cd /e/project/cc-connect-pro/client && go mod tidy`
Expected: 依赖更新成功

- [ ] **Step 3: 运行现有测试**

Run: `cd /e/project/cc-connect-pro/client && go test -v ./...`
Expected: 所有测试通过

- [ ] **Step 4: Commit**

```bash
cd /e/project/cc-connect-pro
git add client/
git commit -m "feat(client): add system tray with dynamic status, close-to-tray behavior"
```
