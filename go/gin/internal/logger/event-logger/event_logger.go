package evtlogger

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/db/entity"

	"pjt/internal/infra/cache"
	"pjt/internal/logger"
	"sync"
	"time"
)

var evetLogger *EventLogger

var sampleLogMu sync.RWMutex

type EventLogger struct {
	dao   dbhandler.DBHandlerInterface
	cache *cache.Cache

	sampleLogQueue []*entity.SampleLog
	sampleLogCnt   int

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

func NewDBLogger(dao dbhandler.DBHandlerInterface, cache *cache.Cache) (*EventLogger, error) {
	context, cancel := context.WithCancel(context.Background())
	return &EventLogger{
		dao:            dao,
		cache:          cache,
		sampleLogQueue: make([]*entity.SampleLog, 100),
		ctx:            context,
		cancel:         cancel,
	}, nil
}

func Shutdown() {
	evetLogger.cancel()
	evetLogger.wg.Wait()
	logger.Infoln("[DB] DB Log Manager goroutine terminated")
}

func StartLogManager() {
	evetLogger.wg.Add(1)
	ticker := time.NewTicker(10 * time.Second)
	defer func() {
		evetLogger.wg.Done()
		ticker.Stop()
	}()

	for {
		select {
		case <-evetLogger.ctx.Done():
			return
		case <-ticker.C:
			sampleLogMu.Lock()
			if evetLogger.sampleLogCnt > 0 {
				for i := 0; i < evetLogger.sampleLogCnt; i++ {

				}
				evetLogger.sampleLogCnt = 0
				go func() {

				}()
			}
			sampleLogMu.Unlock()
		}
	}
}

func SetEventLogger(em *EventLogger) {
	if evetLogger != nil {
		return
	}
	evetLogger = em
}

func PushSampleLog(eventLog *entity.SampleLog, eventLevel int) {
	sampleLogMu.Lock()
	defer sampleLogMu.Unlock()
	if evetLogger.sampleLogCnt < len(evetLogger.sampleLogQueue) {
		evetLogger.sampleLogQueue[evetLogger.sampleLogCnt] = eventLog
	} else {
		evetLogger.sampleLogQueue = append(evetLogger.sampleLogQueue, eventLog)
	}
	evetLogger.sampleLogCnt++
}
