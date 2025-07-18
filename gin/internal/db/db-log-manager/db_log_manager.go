package dblogmanager

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/db/entity"
	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"sync"
	"time"
)

var dbLogManager *DBLogManager

var sampleLogMu sync.RWMutex

type DBLogManager struct {
	dao      dbhandler.DBHandlerInterface
	ram      *ram.Ram
	eventBus *eventbus.EventBus

	sampleLogQueue []*entity.SampleLog
	sampleLogCnt   int

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

func NewDBLogManager(dao dbhandler.DBHandlerInterface, ram *ram.Ram, eventBus *eventbus.EventBus) (*DBLogManager, error) {
	context, cancel := context.WithCancel(context.Background())
	return &DBLogManager{
		dao:            dao,
		ram:            ram,
		eventBus:       eventBus,
		sampleLogQueue: make([]*entity.SampleLog, 100),
		ctx:            context,
		cancel:         cancel,
	}, nil
}

func Shutdown() {
	dbLogManager.cancel()
	dbLogManager.wg.Wait()
	logger.Println("[DB] DB Log Manager goroutine terminated")
}

func StartLogManager() {
	dbLogManager.wg.Add(1)
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		dbLogManager.wg.Done()
		ticker.Stop()
	}()

	for {
		select {
		case <-dbLogManager.ctx.Done():
			return
		case <-ticker.C:
			sampleLogMu.Lock()
			if dbLogManager.sampleLogCnt > 0 {
				for i := 0; i < dbLogManager.sampleLogCnt; i++ {

				}
				dbLogManager.sampleLogCnt = 0
				go func() {

				}()
			}
			sampleLogMu.Unlock()
		}
	}
}

func SetEventManger(em *DBLogManager) {
	if dbLogManager != nil {
		return
	}
	dbLogManager = em
}

func PushSampleLog(log *entity.SampleLog) {
	sampleLogMu.Lock()
	defer sampleLogMu.Unlock()
	if dbLogManager.sampleLogCnt < len(dbLogManager.sampleLogQueue) {
		dbLogManager.sampleLogQueue[dbLogManager.sampleLogCnt] = log
	} else {
		dbLogManager.sampleLogQueue = append(dbLogManager.sampleLogQueue, log)
	}
	dbLogManager.sampleLogCnt++
}
