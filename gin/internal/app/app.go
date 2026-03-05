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

	a.wg.Add(1)
	startRestServer := func() error {
		defer a.wg.Done()
		if err := a.restServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Infof("Failed to start application:\n\t%v", err)
			return err
		}
		return nil
	}

	a.wg.Add(1)
	startTCPServer := func() error {
		defer a.wg.Done()
		if err := a.tcpServer.Start(); err != nil {
			logger.Infof("Failed to start TCP routine:\n\t%v", err)
			return err
		}
		return nil
	}

	// a.wg.Add(1)
	// startSystemMonitoring := func() error {
	// 	defer a.wg.Done()
	// 	if err := a.container.SystemMonitoring.Start(); err != nil {
	// 		logger.Infof("Failed to start SystemMonitoring routine:\n\t%v", err)
	// 		return err
	// 	}
	// 	return nil
	// }

	a.handleShutdown()

	// 30일이 지난 로그 파일 정리
	go logger.StartCleaning()

	/**
	 * DB 로그 루틴 실행
	 * DB 로그 매니저에서 루틴을 또 분기해야할 수도
	 */
	go eventlog.StartLogManager()

	/**
	 * Worker들 실행
	 */
	go a.container.CronWorker.Start()
	go a.container.CacheWorker.Start()
	go a.container.TCPWorker.Start()

	go startTCPServer()

	/**
	 * If managing monitoring thread individually, use the function below, otherwise manage them in system monitoring.
	 */
	// // 5초마다 TCP Listener 상태를 확인, 리스너가 닫혀있다면 다시 리스닝 시도
	// a.tcpServer.StartTCPServerHeartbeat()
	// // 10초마다 DB 연결 상태를 확인, 연결이 끊겨있다면 재연결 시도
	// a.container.Dao.StartDBHeartbeat()

	//go startSystemMonitoring()

	go startRestServer()

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

// monitoring ( routine ) -> tcp ( routine ) -> rest ( routine ) -> log ( routine ) -> each object ( close )
func (a *Application) Shutdown() {

	// shutdown monitoring routine
	a.container.SystemMonitoring.Shutdown()

	// shutdown rest routine
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.restServer.Shutdown(ctx); err != nil {
		logger.Infof("Error during shutdown: %v", err)
	}

	// shutdown db log manager
	eventlog.Shutdown()

	// stop workers job routine
	a.container.CronWorker.Shutdown()
	a.container.CacheWorker.Shutdown()
	a.container.TCPWorker.Shutdown()

	// shutdown tcp routine and tcp heartbeat routine
	a.tcpServer.Shutdown()

	// close servcie object, close: api -> close: dao -> db ctx cancel -> shutdown: DBHeartbeat
	err := a.container.ApiService.Close()
	if err != nil {
		logger.Infof("Failed to close APIService instance:\n\t%v", err)
	}
	a.container.SseManager.Close()

	// EventBus 채널 닫음
	a.container.EventBus.Close()

	if err := redis.CloseRedisClient(); err != nil {
		logger.Infof("Failed to close reids client: %v", err)
	}

	a.wg.Wait()
	logger.Infoln("Application terminated")

	// shutdown log routine
	logger.Shutdown()

	// shutdwon main routine
	a.mainCtxCancel()
}
