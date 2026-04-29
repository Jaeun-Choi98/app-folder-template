package maria

import (
	"context"
	"database/sql"
	"pjt/internal/logger"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Maria struct {
	dsn string
	db  *sql.DB

	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
	heartbeat time.Duration
}

func NewMaria(dsn string, heartbeat time.Duration) (*Maria, error) {
	ctx, cancel := context.WithCancel(context.Background())
	maria := &Maria{
		dsn:       dsn,
		ctx:       ctx,
		cancel:    cancel,
		heartbeat: heartbeat,
		timeout:   5 * time.Second,
	}
	if err := maria.Connect(); err != nil {
		return nil, err
	}
	return maria, nil
}

func (m *Maria) Connect() error {
	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		logger.Infoln(err)
		return err
	}

	// if err := db.Ping(); err != nil {
	// 	logger.Infoln(err)
	// 	db.Close()
	// 	return err
	// }

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	m.db = db
	return nil
}

func (m *Maria) Test() string {
	return "hello world"
}

func (m *Maria) Ping() error {
	if m.db != nil {
		return m.db.Ping()
	}
	return nil
}

func (m *Maria) Close() error {
	m.cancel()
	if m.db != nil {
		return m.db.Close()
	}
	m.wg.Wait()
	return nil
}

/**
 * 일정 시간마다 DB 커넥션을 확인하고, 연결 끊김 시에 재연결을 위한 고루틴 함수
 * Log prefix: [Heartbeat]
 *
 * 아래 함수들은 개별적으로 모니터링이 필요할 때 사용
 * 현재는 사용하지 않고, SystemMonitoring 객체를 이용함
 * 로직은 SystemMonitoring과 같음
 */

func (m *Maria) StartDBHeartbeat() {
	m.wg.Add(1)
	go func() {
		heartbeat := time.NewTicker(m.heartbeat)
		defer func() {
			heartbeat.Stop()
			m.wg.Done()
		}()

		for {
			select {
			case <-m.ctx.Done():
				logger.Infoln("[DB Heartbeat] DB heartbeat goroutine terminated")
				return
			case <-heartbeat.C:
				if err := m.Ping(); err != nil {
					// 재연결 시도
					logger.Infoln("[DB Heartbeat] DB Connection is closed, attempting to reconnect...")
					if err := m.Connect(); err == nil {
						logger.Infoln("[DB Heartbeat] DB reconnection successful")
					}
				}
			}
		}
	}()
}
