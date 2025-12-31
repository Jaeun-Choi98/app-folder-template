package worker

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/cache"
	"pjt/internal/logger"
	"sync"
	"time"
)

/**
 * 1. Cache 데이터를 추적함.
 * 2. 세션 간의 데이터를 비교함. ( 프론트 사용자에 의한 세션과 하드웨어에 의한 세션을 비교 )
 * 3. 필요시 따로 데이터를 가공하여 DB에 저장함.
 *
 * 1,2,3 의 경우, 필요시 DB 이벤트 로그를 남김.
 */
type CacheWorker struct {
	dao   dbhandler.DBHandlerInterface
	cache *cache.Cache
	// eventbus *eventbus.EventBus

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewCacheWorker(d dbhandler.DBHandlerInterface, c *cache.Cache) *CacheWorker {

	ctx, cancel := context.WithCancel(context.Background())

	return &CacheWorker{
		ctx:    ctx,
		cancel: cancel,
		dao:    d,
		cache:  c,
	}
}

func (w *CacheWorker) Start() error {
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
			models, _ := w.cache.GetOprtModels(nil, true)
			if len(models) == 0 {

			}
			// debug
			// logger.Println("here sync DB")

			go func() {
				if err := w.dao.Ping(); err != nil {
					logger.Println("[Cache] Failed to sync DB")
				}
			}()
		}
	}
}

func (w *CacheWorker) Shutdown() error {
	w.cancel()
	w.wg.Wait()
	logger.Printf("[Cache Worker] Cache Worker goroutine is terminated")
	return nil
}
