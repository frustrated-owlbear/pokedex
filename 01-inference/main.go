package main

import (
	"embed"

	"github.com/frustrated-owlbear/pokedex/01-inference/internal/config"
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

	// Create an instance of the app structure
	app := NewApp(cfg)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "01-inference",
		Width:  1024,
		Height: 680,
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
