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
	isPingToDB      bool
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
	healthCheck := time.NewTicker(s.heartbeat * 10)
	var dbComm, tcpListener int8

	recoverProcessTCP := func() {
		logger.Println("[System Monitoring] Listening is closed, attempting to listen...")
		go func() {
			if s.isRecoveringTCP {
				return
			}
			s.isRecoveringTCP = true
			if err := s.Tcp.Listening(); err != nil {
				logger.Printf("[System Monitoring] Failed to listen:\n\t%v", err)
			} else {
				logger.Println("[System Monitoring] TCP listening successful")
			}
			s.isRecoveringTCP = false
		}()
	}

	recoverProcessDB := func() {
		logger.Println("[System Monitoring] DB Connection is closed, attempting to reconnect...")

		go func() {
			if s.isRecoveringDB {
				return
			}
			s.isRecoveringDB = true
			if err := s.Dao.Connect(); err == nil {
				logger.Println("[System Monitoring] DB reconnection successful")
			}
			s.isRecoveringDB = false
		}()
	}

	defer func() {
		s.wg.Done()
		healthCheck.Stop()
		heartbeat.Stop()
	}()

	for {
		select {
		case <-s.ctx.Done():
			logger.Println("[System Monitoring] System Monitoring goroutine terminated")
			return nil
		case <-heartbeat.C:
			if !s.Tcp.CheckConnection() {
				recoverProcessTCP()
				tcpListener = 0
			} else {
				tcpListener = 1
			}

			if !s.isPingToDB {
				go func() {
					s.isPingToDB = true
					if err := s.Dao.Ping(); err != nil {
						recoverProcessDB()
						dbComm = 0
					} else {
						dbComm = 1
					}
					s.isPingToDB = false
				}()
			}

		case <-healthCheck.C:
			if s.Tcp.IsListening() && !s.Tcp.CheckConnection() {
				recoverProcessTCP()
				tcpListener = 0
			} else {
				tcpListener = 1
			}

			if !s.isPingToDB {
				go func() {
					s.isPingToDB = true
					if err := s.Dao.Ping(); err != nil {
						recoverProcessDB()
						dbComm = 0
					} else {
						dbComm = 1
					}
					s.isPingToDB = false
				}()
			}
		}

		// 이벤트 버스에 데이터 publish
		s.EventBus.Publish(eventbus.SysSttType, eventbus.SysStt.Add(eventbus.SysSttPayload{
			ServerTime:  time.Now().Format(time.DateTime),
			DBState:     dbComm,
			TCPListener: tcpListener,
		}))

	}
}

func (s *SystemMonitoring) Shutdown() error {
	s.cancel()
	s.wg.Wait()
	return nil
}
