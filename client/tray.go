package main

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// setupTray creates and configures the system tray icon with a context menu.
func setupTray(app *application.App, window *application.WebviewWindow, serviceApp *App) (*application.SystemTray, *application.Menu) {
	tray := app.SystemTray.New()

	tray.SetTooltip("cc-connect")
	tray.SetIcon(defaultIconBytes)

	menu := application.NewMenu()

	// Status item (disabled, shows current service state)
	statusItem := menu.Add("状态: 未启动")
	statusItem.SetEnabled(false)

	menu.AddSeparator()

	// Show window
	showItem := menu.Add("显示窗口")
	showItem.OnClick(func(*application.Context) {
		window.Show()
		window.Focus()
	})

	// Restart service
	restartItem := menu.Add("重启服务")
	restartItem.OnClick(func(*application.Context) {
		go func() {
			_ = serviceApp.RestartService()
		}()
	})

	menu.AddSeparator()

	// Quit
	quitItem := menu.Add("退出")
	quitItem.OnClick(func(*application.Context) {
		_ = serviceApp.Shutdown()
		application.Get().Quit()
	})

	tray.SetMenu(menu)

	// Click tray icon to show window
	tray.OnClick(func() {
		window.Show()
		window.Focus()
	})

	// Store references on serviceApp for later updates
	serviceApp.tray = tray
	serviceApp.trayMenu = menu
	serviceApp.statusItem = statusItem

	return tray, menu
}

// updateTrayStatus updates the tray icon and status menu item based on the service status.
func updateTrayStatus(tray *application.SystemTray, menu *application.Menu, status ServiceStatus) {
	var label string
	var icon []byte

	switch status {
	case StatusIdle:
		label = "状态: 已停止"
		icon = defaultIconBytes
	case StatusStarting:
		label = "状态: 启动中..."
		icon = defaultIconBytes
	case StatusRunning:
		label = "状态: 运行中"
		icon = accentIconBytes
	case StatusStopping:
		label = "状态: 停止中..."
		icon = defaultIconBytes
	case StatusError:
		label = "状态: 错误"
		icon = defaultIconBytes
	default:
		label = fmt.Sprintf("状态: %s", status)
		icon = defaultIconBytes
	}

	tray.SetIcon(icon)
	tray.SetTooltip("cc-connect - " + label)

	// Update the status menu item label
	statusItem := menu.FindByLabel("状态: 未启动")
	if statusItem == nil {
		// Try to find by partial match using ItemAt(0)
		statusItem = menu.ItemAt(0)
	}
	if statusItem != nil {
		statusItem.SetLabel(label)
	}
}
