package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/app"
	"pjt/internal/container"
	"pjt/internal/logger"
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

	// 30일이 지난 로그 파일 정리
	logger.StartCleaning()

	// 10초마다 DB 연결 상태를 확인, 연결이 끊겨있다면 재연결 시도
	container.Dao.StartDBHeartbeat()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			container.Service.Close()
			if err != nil {
				logger.Printf("Failed to Close Service Instance:\n\t%v", err)
			}
			container.Dao.StopDBHeartbeat()
			logger.StopCleaning()
			wg.Done()
		}()

		if err := app.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Failed to start application:\n\t%v", err)
			return
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	logger.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Printf("Error during shutdown: %v", err)
		return
	}

	wg.Wait()
	logger.Println("Application terminated")
}
