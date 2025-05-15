package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/app"
	"pjt/internal/container"
	"sync"
	"syscall"
	"time"
)

func main() {
	container, err := container.NewContainer()
	if err != nil {
		return
	}
	app := app.NewApplication(container)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
