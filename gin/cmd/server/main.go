package main

import (
	"context"
	"pjt/internal/app"
	"pjt/internal/container"
	"pjt/internal/logger"
	"time"
)

func main() {
	container, err := container.NewContainer()
	if err != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	app := app.NewApplication(container, cancel)

	if err := app.Start(); err != nil {
		logger.Printf("Failed to start Application:\n\t%v", err)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			container.TCPServer.HeartbeatTest()
		}
	}

}
