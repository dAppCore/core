package main

import (
	"embed"
	"fmt"

	"core"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Services: []application.Service{
			application.NewService(core.New(assets)),
		},
	})

	core.Setup(app)

	app.Event.On("setup-done", func(e *application.CustomEvent) {
		fmt.Println("Setup done!")
	})

	err := app.Run()
	if err != nil {
		panic(err)
	}
}
