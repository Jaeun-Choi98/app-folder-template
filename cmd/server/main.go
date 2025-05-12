package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"root/internal/app"
	"root/internal/container"
	"sync"
	"syscall"
	"time"
)

func main() {
	container := container.NewContainer()

	app := app.NewApplication(container)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(); err != nil {
			log.Printf("Failed to start application: %v", err)
			return
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}

	wg.Wait()
	log.Println("Application stopped")
}
