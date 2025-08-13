package main

import (
	"log"

	"simple-tui-cal/internal/ui"
)

func main() {
	app, err := ui.NewApp()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
