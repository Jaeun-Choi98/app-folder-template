package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/container"
	"pjt/internal/logger"
	rest "pjt/internal/transport/http-rest"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	container  *container.Container
	restServer *rest.RESTServer
	wg         *sync.WaitGroup
	cancel     context.CancelFunc
}

func NewApplication(container *container.Container, cancel context.CancelFunc) *Application {

	return &Application{
		container:  container,
		wg:         &sync.WaitGroup{},
		restServer: container.RESTServer,
		cancel:     cancel,
	}
}

func (a *Application) Init() error {

	// 30일이 지난 로그 파일 정리
	logger.StartCleaning()

	// 10초마다 DB 연결 상태를 확인, 연결이 끊겨있다면 재연결 시도
	a.container.Dao.StartDBHeartbeat()

	return nil
}

func (a *Application) Start() error {
	a.wg.Add(1)

	startRestServer := func() error {
		defer a.wg.Done()

		if err := a.restServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Failed to start application:\n\t%v", err)
			return err
		}

		return nil
	}
	a.handleShutdown()
	return startRestServer()
}

func (a *Application) handleShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Println("Shutting down...")
		a.Shutdown()
	}()
}

func (a *Application) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.restServer.Shutdown(ctx); err != nil {
		logger.Printf("Error during shutdown: %v", err)
		return err
	}

	err := a.container.Service.Close()
	if err != nil {
		logger.Printf("Failed to Close Service Instance:\n\t%v", err)
	}
	// db healthcheck 고루틴 종료
	a.container.Dao.StopDBHeartbeat()
	// 로그 정리 고루틴 종료
	logger.StopCleaning()
	// 로그 파일 닫음
	logger.Close()

	a.wg.Wait()
	logger.Println("Application terminated")
	a.cancel()
	return nil
}
