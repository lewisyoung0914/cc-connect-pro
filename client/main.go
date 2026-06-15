package main

import (
	"embed"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
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

	// Create the serviceApp instance before RegisterService so we can store
	// window and tray references on it.
	serviceApp := &App{}

	app.RegisterService(application.NewService(serviceApp))

	// Create the main window and store the reference on serviceApp.
	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "cc-connect",
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

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
