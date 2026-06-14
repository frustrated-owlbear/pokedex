package main

import (
	"embed"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		println("Error loading config:", err.Error())
		return
	}

	app, err := NewApp(cfg)
	if err != nil {
		println("Error creating app:", err.Error())
		return
	}

	err = wails.Run(&options.App{
		Title:  "05-session-management",
		Width:  1280,
		Height: 840,
		MinWidth: 960,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 239, G: 239, B: 235, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
