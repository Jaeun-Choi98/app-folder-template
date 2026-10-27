package main

import (
	"context"
	"log"
	"pjt/internal/app"
	"pjt/internal/config"
	"pjt/internal/container"
	"time"
)

// func main() {

// 	container, err := container.NewContainer()
// 	if err != nil {
// 		logger.Infoln(err)
// 		return
// 	}
// 	ctx, cancel := context.WithCancel(context.Background())
// 	app := app.NewApplication(container, cancel)

// 	app.Start()

// 	ticker := time.NewTicker(7 * time.Second)
// 	defer ticker.Stop()
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			//container.TCPServer.HeartbeatTest()
// 		}
// 	}

// }
func main() {

	cfg, err := config.NewConfiguration()
	if err != nil {
		log.Printf("[Main] failed to create new configuration, err: %v", err)
		return
	}

	switch cfg.DeployMode {
	case config.TEST:
		Test(cfg)
	case config.PRODUCT:
		Product(cfg)
	default:
		log.Printf("[Main] undefined deploy stage, deploy mode: %v", cfg.DeployMode)
		return
	}
}

func Product(cfg *config.Configuration) {
	for {
		if err := container.NewContainer(cfg); err != nil {
			log.Println(err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		app := app.NewApplication(cancel)

		app.Start()
		<-ctx.Done()

		container.DelContainer()

		log.Printf("[Main] main process terminated. restarting in 3 seconds.")

		time.Sleep(3 * time.Second)
	}
}

func Test(cfg *config.Configuration) {
	if err := container.NewContainer(cfg); err != nil {
		log.Println(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	app := app.NewApplication(cancel)

	app.Start()
	<-ctx.Done()

	container.DelContainer()
}
