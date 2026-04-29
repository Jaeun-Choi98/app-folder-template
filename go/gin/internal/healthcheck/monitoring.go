package monitoring

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/event"
	"pjt/internal/logger"

	tcp "pjt/internal/transport/tcp/server"
	"sync"
	"time"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type SystemMonitoring struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	heartbeat       time.Duration
	isPingToDB      bool
	isRecoveringDB  bool
	isRecoveringTCP bool

	Dao dbhandler.DBHandlerInterface
	Tcp *tcp.TCPServer
}

func NewSystemMonitoring(dao dbhandler.DBHandlerInterface, tcp *tcp.TCPServer, heartbeat time.Duration) *SystemMonitoring {
	ctx, cancel := context.WithCancel(context.Background())
	return &SystemMonitoring{
		ctx:       ctx,
		cancel:    cancel,
		heartbeat: heartbeat,
		Dao:       dao,
		Tcp:       tcp,
	}
}

func (s *SystemMonitoring) Start() error {

	s.wg.Add(1)
	heartbeat := time.NewTicker(s.heartbeat)
	healthCheck := time.NewTicker(s.heartbeat * 10)
	var dbComm, tcpListener int8

	recoverProcessTCP := func() {
		logger.Infoln("[System Monitoring] Listening is closed, attempting to listen...")
		go func() {
			if s.isRecoveringTCP {
				return
			}
			s.isRecoveringTCP = true
			if err := s.Tcp.Listening(); err != nil {
				logger.Infof("[System Monitoring] Failed to listen:\n\t%v", err)
			} else {
				logger.Infoln("[System Monitoring] TCP listening successful")
			}
			s.isRecoveringTCP = false
		}()
	}

	recoverProcessDB := func() {
		logger.Infoln("[System Monitoring] DB Connection is closed, attempting to reconnect...")

		go func() {
			if s.isRecoveringDB {
				return
			}
			s.isRecoveringDB = true
			if err := s.Dao.Connect(); err == nil {
				logger.Infoln("[System Monitoring] DB reconnection successful")
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
			logger.Infoln("[System Monitoring] System Monitoring goroutine terminated")
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

		// sse로 전달
		eventbus.Publish(event.GetEventBus(), "sse.event",
			event.NewSseEventSysStt(
				&event.SysStt{
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
