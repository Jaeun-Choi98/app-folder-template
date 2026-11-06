package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/container"
	"pjt/internal/event"
	eventlog "pjt/internal/logger/event-logger"
	"pjt/internal/redis"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Jaeun-Choi98/modules/logger"
)

var (
	ErrSignal = errors.New("terminated by signal")
	ErrPanic  = errors.New("terminated by panic")
)

type Application struct {
	wg            sync.WaitGroup
	mainCtxCancel context.CancelCauseFunc

	stopOnce  sync.Once
	signalErr atomic.Pointer[error]
	done      chan struct{}
}

func NewApplication(cancel context.CancelCauseFunc) *Application {
	return &Application{
		mainCtxCancel: cancel,
		done:          make(chan struct{}),
	}
}

func (a *Application) safeGo(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[App][%s] panic: %v\n%s", name, r, debug.Stack())
				a.stop(fmt.Errorf("%w: [%s] %v", ErrPanic, name, r))
			}
		}()
		fn()
	}()
}

func (a *Application) Start() {
	a.handleSignal()

	a.safeGo("log-manager", func() { eventlog.StartLogManager() })
	a.safeGo("cron-worker", func() { container.GetCronWorker().Start() })
	a.safeGo("cache-worker", func() { container.GetCacheWorker().Start() })

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

	a.safeGo("system-monitoring", func() { container.GetSystemMonitoring().Start() })
}

func (a *Application) handleSignal() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(signalChan)

		select {
		case sig := <-signalChan:
			switch sig {
			case syscall.SIGINT:
				logger.Infof("[App] user use ctrl+c(signal: %v)", sig)
			case syscall.SIGTERM:
				logger.Infof("[App] received exit(signal: %v)", sig)
			}
			a.stop(fmt.Errorf("%w: %v", ErrSignal, sig))
		case <-a.done:
		}
	}()
}

// stop 은 시그널/panic 어느 쪽에서 불려도 정리를 한 번만 수행하고, 끝나면 원인을 main 에 넘긴다.
// panic 정리 도중 시그널이 들어오면 그 호출은 정리가 끝날 때까지 기다리고, 원인은 시그널로 바뀐다.
func (a *Application) stop(cause error) {
	if errors.Is(cause, ErrSignal) {
		a.signalErr.CompareAndSwap(nil, &cause)
	}

	a.stopOnce.Do(func() {
		logger.Infof("[App] Shutting down... (cause: %v)", cause)
		a.shutdown()

		if sigErr := a.signalErr.Load(); sigErr != nil {
			cause = *sigErr
		}
		close(a.done)
		a.mainCtxCancel(cause)
	})
}

// Shutdown order: SystemMonitoring → REST → DB log manager → Workers → TCP → services → EventBus → Redis → logger.
// Reversing this order risks panics.
func (a *Application) shutdown() {
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

	if err := logger.Close(); err != nil {
		log.Printf("[App] Failed to close logger: %v", err)
	}
}
