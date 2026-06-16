package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	serviceApp := &App{}
	serviceApp.status = StatusIdle

	serviceApp.service = NewService(serviceApp)
	serviceApp.logBuffer = NewLogRingBuffer(500)
	// NOTE: setupLogCapture is NOT called here! It is lazily installed
	// inside App.GetRecentLogs() on the first frontend request for logs.
	// Replacing slog.Default() before app.Run() interferes with Wails v3's
	// internal logging pipeline — Wails uses slog for WebView2/window
	// lifecycle messages, and wrapping its handler causes the window to
	// silently fail to appear on Windows 11 (alpha.98).

	app := application.New(application.Options{
		Name:        "cc-connect-pro",
		Description: "cc-connect-pro Desktop Client",
		Services: []application.Service{
			application.NewService(serviceApp),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "cc-connect-pro",
		Width:  960,
		Height: 640,
		URL:    "/",
	})
	serviceApp.window = window

	// Intercept window close: hide instead of quit so the app runs in tray.
	window.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		event.Cancel()
		window.Hide()
	})

	// Setup system tray with context menu.
	setupTray(app, window, serviceApp)

	// Register service status event listeners.
	app.Event.On("service:starting", func(event *application.CustomEvent) {
		serviceApp.status = StatusStarting
		if serviceApp.tray != nil {
			updateTrayStatus(serviceApp, serviceApp.status)
		}
	})
	app.Event.On("service:running", func(event *application.CustomEvent) {
		serviceApp.status = StatusRunning
		if serviceApp.tray != nil {
			updateTrayStatus(serviceApp, serviceApp.status)
		}
	})
	app.Event.On("service:stopping", func(event *application.CustomEvent) {
		serviceApp.status = StatusStopping
		if serviceApp.tray != nil {
			updateTrayStatus(serviceApp, serviceApp.status)
		}
	})
	app.Event.On("service:idle", func(event *application.CustomEvent) {
		serviceApp.status = StatusIdle
		if serviceApp.tray != nil {
			updateTrayStatus(serviceApp, serviceApp.status)
		}
	})
	app.Event.On("service:error", func(event *application.CustomEvent) {
		serviceApp.status = StatusError
		if serviceApp.tray != nil {
			updateTrayStatus(serviceApp, serviceApp.status)
		}
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
