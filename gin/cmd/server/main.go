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
		logger.Println(err)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	app := app.NewApplication(container, cancel)

	app.Start()

	ticker := time.NewTicker(7 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			//container.TCPServer.HeartbeatTest()
		}
	}

}
