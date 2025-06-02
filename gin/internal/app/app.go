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
	tcp "pjt/internal/transport/tcp/server"
	"sync"
	"syscall"
	"time"
)

type Application struct {
	container     *container.Container
	restServer    *rest.RESTServer
	tcpServer     *tcp.TCPServer
	wg            sync.WaitGroup
	mainCtxCancel context.CancelFunc
}

func NewApplication(container *container.Container, cancel context.CancelFunc) *Application {

	return &Application{
		container:     container,
		restServer:    container.RESTServer,
		tcpServer:     container.TCPServer,
		mainCtxCancel: cancel,
	}
}

func (a *Application) Init() error {
	return nil
}

func (a *Application) Start() {

	// 30일이 지난 로그 파일 정리
	logger.StartCleaning()

	// 10초마다 DB 연결 상태를 확인, 연결이 끊겨있다면 재연결 시도
	a.container.Dao.StartDBHeartbeat()

	a.wg.Add(1)
	startRestServer := func() error {
		defer a.wg.Done()
		if err := a.restServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Printf("Failed to start application:\n\t%v", err)
			return err
		}
		return nil
	}

	a.wg.Add(1)
	startTCPServer := func() error {
		defer a.wg.Done()
		if err := a.tcpServer.Start(); err != nil {
			logger.Printf("Failed to start TCP routine:\n\t%v", err)
			return err
		}
		return nil
	}

	a.handleShutdown()

	go startTCPServer()

	a.tcpServer.StartTCPServerHeartbeat()

	go startRestServer()

}

func (a *Application) handleShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		logger.Println("Shutting down...")
		a.Shutdown()
		close(signalChan)
	}()
}

func (a *Application) Shutdown() {

	// shutdown tcp routine and tcp heartbeat routine
	a.tcpServer.Shutdown()

	// shutdown rest routine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.restServer.Shutdown(ctx); err != nil {
		logger.Printf("Error during shutdown: %v", err)
	}

	err := a.container.ApiService.Close()
	if err != nil {
		logger.Printf("Failed to Close Service Instance:\n\t%v", err)
	}
	a.container.SseService.Close()

	// db healthcheck 고루틴 종료
	a.container.Dao.StopDBHeartbeat()
	// 로그 정리 고루틴 종료
	logger.StopCleaning()
	// 로그 파일 닫음
	logger.Close()
	// 컨테이너 객체 nil로 초기화
	a.container.Close()

	a.wg.Wait()
	logger.Println("Application terminated")

	// shutdwon main routine
	a.mainCtxCancel()
}
