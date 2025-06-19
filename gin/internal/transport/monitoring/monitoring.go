package monitoring

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	tcp "pjt/internal/transport/tcp/server"
	"sync"
	"time"
)

type SystemMonitoring struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	heartbeat       time.Duration
	isRecoveringDB  bool
	isRecoveringTCP bool

	Dao      dbhandler.DBHandlerInterface
	Tcp      *tcp.TCPServer
	EventBus *eventbus.EventBus
}

func NewSystemMonitoring(dao dbhandler.DBHandlerInterface, tcp *tcp.TCPServer, eventBus *eventbus.EventBus, heartbeat time.Duration) *SystemMonitoring {
	ctx, cancel := context.WithCancel(context.Background())
	return &SystemMonitoring{
		ctx:       ctx,
		cancel:    cancel,
		heartbeat: heartbeat,
		Dao:       dao,
		Tcp:       tcp,
		EventBus:  eventBus,
	}
}

func (s *SystemMonitoring) Start() error {

	s.wg.Add(1)
	heartbeat := time.NewTicker(s.heartbeat)
	var dbComm, tcpListener int8

	defer func() {
		s.wg.Done()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-s.ctx.Done():
			logger.Println("[System Monitoring] System Monitoring goroutine terminated")
			return nil
		case <-heartbeat.C:
			if !s.Tcp.CheckConnection() || !s.Tcp.IsListening() {
				logger.Println("[System Monitoring] Listening is closed, attempting to reconnect...")
				go func() {
					if s.isRecoveringTCP {
						logger.Println("tcp 회복중입니다.")
						return
					}
					s.isRecoveringTCP = true
					if err := s.Tcp.Listening(); err != nil {
						logger.Printf("[System Monitoring] Failed to listen:\n\t%v", err)
					}
					s.isRecoveringTCP = false
				}()
				tcpListener = 0
			} else {
				tcpListener = 1
			}

			if err := s.Dao.Ping(); err != nil {
				// 재연결 시도
				logger.Println("[DB Heartbeat] DB Connection is closed, attempting to reconnect...")

				go func() {
					if s.isRecoveringDB {
						return
					}
					s.isRecoveringDB = true
					if err := s.Dao.Connect(); err == nil {
						logger.Println("[DB Heartbeat] DB reconnection successful")
					}
					s.isRecoveringDB = false
				}()

				dbComm = 0
			} else {
				dbComm = 1
			}
		}

		// 이벤트 버스에 데이터 publish
		s.EventBus.Publish(eventbus.SysSttType,
			&eventbus.SysStt{Type: "sysstt",
				Payload: map[string]any{"svrTime": time.Now().Format(time.DateTime),
					"dbComm": dbComm, "tcpListenr": tcpListener}})
	}
}

func (s *SystemMonitoring) Shutdown() error {
	s.cancel()
	s.wg.Wait()
	return nil
}
