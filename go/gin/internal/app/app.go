package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/container"
	"pjt/internal/logger"
	eventlog "pjt/internal/logger/event-logger"
	"pjt/internal/redis"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	container     *container.Container
	wg            sync.WaitGroup
	mainCtxCancel context.CancelFunc
}

func NewApplication(c *container.Container, cancel context.CancelFunc) *Application {
	return &Application{
		container:     c,
		mainCtxCancel: cancel,
	}
}

func (a *Application) Start() {
	a.handleShutdown()

	go logger.StartCleaning()
	go eventlog.StartLogManager()
	go a.container.CronWorker.Start()
	go a.container.CacheWorker.Start()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.container.TCPServer.Start(); err != nil {
			logger.Infof("Failed to start TCP server:\n\t%v", err)
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := a.container.RESTServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Failed to start REST server:\n\t%v", err)
		}
	}()
}

func (a *Application) handleShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Infoln("Shutting down...")
		a.Shutdown()
		close(signalChan)
	}()
}

// Shutdown order: SystemMonitoring → REST → DB log manager → Workers → TCP → services → EventBus → Redis → logger.
// Reversing this order risks panics.
func (a *Application) Shutdown() {
	a.container.SystemMonitoring.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := a.container.RESTServer.Shutdown(ctx); err != nil {
		logger.Infof("Error during REST shutdown: %v", err)
	}

	eventlog.Shutdown()

	a.container.CronWorker.Shutdown()
	a.container.CacheWorker.Shutdown()

	a.container.TCPClientManager.Close()
	a.container.TCPServer.Shutdown()

	if err := a.container.ApiService.Close(); err != nil {
		logger.Infof("Failed to close APIService:\n\t%v", err)
	}
	a.container.SseManager.Close()

	a.container.EventBus.Close()

	if err := redis.CloseRedisClient(); err != nil {
		logger.Infof("Failed to close Redis client: %v", err)
	}

	a.wg.Wait()
	logger.Infoln("Application terminated")

	logger.Shutdown()
	a.mainCtxCancel()
}
