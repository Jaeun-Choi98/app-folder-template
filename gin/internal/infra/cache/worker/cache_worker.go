package cacheworker

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/cache"
	"pjt/internal/logger"
	"sync"
	"time"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type Worker struct {
	dao      dbhandler.DBHandlerInterface
	eventbus *eventbus.EventBus
	cache    *cache.Cache
	// 정류장 상태 및 장비 상태 동기화를 위한 변수
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewCacheWorker(dao dbhandler.DBHandlerInterface) *Worker {

	ctx, cancel := context.WithCancel(context.Background())

	return &Worker{
		ctx:    ctx,
		cancel: cancel,
		dao:    dao,
	}
}

func (w *Worker) Start() error {
	w.wg.Add(1)
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		w.wg.Done()
	}()
	for {
		select {
		case <-w.ctx.Done():
			return nil
		case <-ticker.C:
			models, _ := w.GetOprtModels(nil, true)
			if len(models) == 0 {

			}
			// debug
			// logger.Println("here sync DB")

			go func() {
				if err := o.dao.Ping(); err != nil {
					logger.Println("[Cache] Failed to sync DB")
				}
			}()
		}
	}
}
