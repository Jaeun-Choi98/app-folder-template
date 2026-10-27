package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/container"
	"pjt/internal/event"
	"pjt/internal/logger"
	eventlog "pjt/internal/logger/event-logger"
	"pjt/internal/redis"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	wg            sync.WaitGroup
	mainCtxCancel context.CancelFunc
}

func NewApplication(cancel context.CancelFunc) *Application {
	return &Application{
		mainCtxCancel: cancel,
	}
}

func (a *Application) safeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Infof("[App][%s] panic: %v\n%s", name, r, debug.Stack())
				a.mainCtxCancel() // main의 <-ctx.Done() 깨움 → 재시작 경로
			}
		}()
		fn()
	}()
}

func (a *Application) Start() {
	a.handleShutdown()

	logger.StartCleaning()
	a.safeGo("log-manager", func() { eventlog.StartLogManager() })
	a.safeGo("cron-worker", func() { container.GetCronWorker().Start() })
	a.safeGo("cron-worker", func() { container.GetCacheWorker().Start() })

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := container.GetTCPServer().Start(); err != nil {
			logger.Infof("[App] Failed to start TCP server:\n\t%v", err)
		}
	}()

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		if err := container.GetRESTServer().Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Infof("[App] Failed to start REST server:\n\t%v", err)
		}
	}()

	a.safeGo("cron-worker", func() { container.GetSystemMonitoring().Start() })
}

func (a *Application) handleShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		signal.Stop(signalChan)
		logger.Infoln("[App] Shutting down...")
		a.Shutdown()
		close(signalChan)
	}()
}

// Shutdown order: SystemMonitoring → REST → DB log manager → Workers → TCP → services → EventBus → Redis → logger.
// Reversing this order risks panics.
func (a *Application) Shutdown() {
	container.GetSystemMonitoring().Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := container.GetRESTServer().Shutdown(ctx); err != nil {
		logger.Infof("[App] Error during REST shutdown: %v", err)
	}

	eventlog.Shutdown()

	container.GetCronWorker().Shutdown()
	container.GetCacheWorker().Shutdown()

	container.GetTcpClientManager().Close()
	container.GetTCPServer().Shutdown()

	if err := container.GetApiService().Close(); err != nil {
		logger.Infof("[App] Failed to close APIService:\n\t%v", err)
	}
	container.GetSseManager().Close()

	event.GetEventBus().Close()

	if err := redis.CloseRedisClient(); err != nil {
		logger.Infof("[App] Failed to close Redis client: %v", err)
	}

	a.wg.Wait()
	logger.Infoln("[App] Application terminated")

	logger.Shutdown()
	a.mainCtxCancel()
}
