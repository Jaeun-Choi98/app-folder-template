package repository

import (
	"context"
	"database/sql"
	"pjt/internal/config"
	"pjt/internal/logger"
	"sync"
	"time"
	//_ "github.com/godror/godror"
)

var heartbeat = time.Second * 10

type Oracle struct {
	wg     *sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker

	config *config.Configuration
	db     *sql.DB
}

func NewOralce(config *config.Configuration) (DaoInterface, error) {
	// connInfo := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, config.User, config.Passwd, config.Conn)
	// db, err := sql.Open("godror", connInfo)
	// if err != nil {
	// 	logger.Println(err)
	// 	return nil, err
	// }

	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(10)

	ctx, cancel := context.WithCancel(context.Background())

	return &Oracle{
		wg:     &sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
		ticker: time.NewTicker(heartbeat),
		db:     nil,
		config: config,
	}, nil
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
	if o.db != nil {
		return o.db.Close()
	}
	return nil
}

/**
 * 일정 시간마다 DB 커넥션을 확인하고, 연결 끊김 시에 재연결을 위한 고루틴 함수
 * Log prefix: [Heartbeat]
 */
func (o *Oracle) StartDBHeartbeat() {
	o.wg.Add(1)
	go func() {
		defer o.wg.Done()
		for {
			select {
			case <-o.ctx.Done():
				logger.Println("[Heartbeat] DB heartbeat goroutine terminated")
				return
			case <-o.ticker.C:
				if err := o.Ping(); err != nil {
					logger.Printf("DB connection loss detected:\n\t%v", err)
					// 재연결 시도
					for i := 0; i < 3; i++ {
						logger.Printf("Reconnecting to DB #%d", i+1)
						if err := o.reConnect(); err == nil {
							logger.Println("DB reconnection successful")
							break
						}
						// 일정 시간 대기 후 재시도
						time.Sleep(5 * time.Second)
					}
				}
			}
		}
	}()
}

func (o *Oracle) reConnect() error {
	// connInfo := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, config.User, config.Passwd, config.Conn)
	// db, err := sql.Open("godror", connInfo)
	// if err != nil {
	// 	logger.Println(err)
	// 	return nil, err
	// }

	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(10)
	o.db = nil
	return nil
}

func (o *Oracle) StopDBHeartbeat() {
	o.cancel()
	o.ticker.Stop()
	o.wg.Wait()
	if o.db != nil {
		o.db.Close()
	}
}
