package main

import (
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"

	"kineticgo/internal/wailsapp"
)

func main() {
	app := wailsapp.InitializeApp()

	err := wails.Run(&options.App{
		Title:      "KineticGo",
		Width:      1280,
		Height:     800,
		MinWidth:   1024,
		MinHeight:  768,
		Assets:     DesktopDist,
		OnStartup:  app.OnStartup,
		OnShutdown: app.OnShutdown,

		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
