package main

import (
	"context"
	"pjt/internal/app"
	"pjt/internal/container"
	"pjt/internal/logger"
)

func main() {
	container, err := container.NewContainer()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())

	app := app.NewApplication(container, cancel)

	app.Init()

	if err := app.Start(); err != nil {
		logger.Printf("Failed to start Application:\n\t%v", err)
		return
	}

	select {
	case <-ctx.Done():
		return
	}
}
