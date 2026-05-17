package main

import (
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"

	"kineticgo/internal/api"
	"kineticgo/internal/wailsapp"
)

func main() {
	app := wailsapp.InitializeApp()

	go func() {
		h := api.NewHandler()
		if err := h.Serve(":1234"); err != nil {
			log.Fatal(err)
		}
	}()

	err := wails.Run(&options.App{
		Title:     "KineticGo",
		Width:     1024,
		Height:    768,
		Assets:    DesktopDist,
		OnStartup: app.OnStartup,

		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
