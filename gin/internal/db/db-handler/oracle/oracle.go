package oracle

import (
	"context"
	"database/sql"
	"pjt/internal/logger"
	"sync"
	"time"

	_ "github.com/godror/godror"
)

type Oracle struct {
	dsn string
	db  *sql.DB

	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	heartbeat time.Duration
}

func NewOralce(dsn string, heartbeat time.Duration) (*Oracle, error) {

	ctx, cancel := context.WithCancel(context.Background())
	oracle := &Oracle{
		ctx:       ctx,
		cancel:    cancel,
		heartbeat: heartbeat,
		dsn:       dsn,
	}

	if err := oracle.Connect(); err != nil {
		return nil, err
	}

	return oracle, nil
}

func (o *Oracle) Connect() error {
	db, err := sql.Open("godror", o.dsn)
	if err != nil {
		logger.Println(err)
		return err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	o.db = db
	return nil
}

func (o *Oracle) Test() string {
	return "hello world"
}

func (o *Oracle) Ping() error {
	if o.db != nil {
		return o.db.Ping()
	}
	return nil
}

func (o *Oracle) Close() error {
	o.cancel()
	if o.db != nil {
		return o.db.Close()
	}
	o.wg.Wait()
	return nil
}

/**
 * 일정 시간마다 DB 커넥션을 확인하고, 연결 끊김 시에 재연결을 위한 고루틴 함수
 * Log prefix: [Heartbeat]
 *  If managing monitoring thread individually, use the function below, otherwise manage them in system monitoring.
 */

func (o *Oracle) StartDBHeartbeat() {
	o.wg.Add(1)
	go func() {
		heartbeat := time.NewTicker(o.heartbeat)
		defer func() {
			heartbeat.Stop()
			o.wg.Done()
		}()

		for {
			select {
			case <-o.ctx.Done():
				logger.Println("[DB Heartbeat] DB heartbeat goroutine terminated")
				return
			case <-heartbeat.C:
				if err := o.Ping(); err != nil {
					// 재연결 시도
					logger.Println("[DB Heartbeat] DB Connection is closed, attempting to reconnect...")
					if err := o.Connect(); err == nil {
						logger.Println("[DB Heartbeat] DB reconnection successful")
					}
				}
			}
		}
	}()
}
