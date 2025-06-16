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
	heartbeat time.Duration
}

func NewMaria(dsn string, heartbeat time.Duration) (*Maria, error) {
	ctx, cancel := context.WithCancel(context.Background())
	maria := &Maria{
		dsn:       dsn,
		ctx:       ctx,
		cancel:    cancel,
		heartbeat: heartbeat,
	}
	if err := maria.Connect(); err != nil {
		return nil, err
	}
	return maria, nil
}

func (m *Maria) Connect() error {
	db, err := sql.Open("mysql", m.dsn)
	if err != nil {
		logger.Println(err)
		return err
	}

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
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

/**
 * 일정 시간마다 DB 커넥션을 확인하고, 연결 끊김 시에 재연결을 위한 고루틴 함수
 * Log prefix: [Heartbeat]
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
				logger.Println("[DB Heartbeat] DB heartbeat goroutine terminated")
				return
			case <-heartbeat.C:
				if err := m.Ping(); err != nil {
					// 재연결 시도
					logger.Println("[DB Heartbeat] DB Connection is closed, attempting to reconnect...")
					if err := m.Connect(); err == nil {
						logger.Println("[DB Heartbeat] DB reconnection successful")
					}
				}
			}
		}
	}()
}

func (m *Maria) StopDBHeartbeat() {
	m.cancel()
	m.wg.Wait()
	if m.db != nil {
		m.db.Close()
	}
}
