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

	app.RegisterService(application.NewService(&App{}))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
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
