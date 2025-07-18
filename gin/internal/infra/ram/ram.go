package ram

import (
	"context"
	"errors"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"pjt/internal/logger"
	"sync"
	"time"
)

/**
 * this layer is role of in memory db.
 * it needs reloading on start or modifing resource ( delete(both logical/physical), insert not update )
 *
 * when operating on get method. ram.cache -> service or other layer:
 * make new instance using mutex. it also convenient to update object
 * should provice pk filter, unique map / should except a logical deleted records. ( provide view data )
 */

var (
	errNotExistId = errors.New("not exists id ( some or one )")
)

type Ram struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int64]*object.Sample

	// 정류장 상태 및 장비 상태 동기화를 위한 변수
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewRam(dao dbhandler.DBHandlerInterface) (*Ram, error) {

	ctx, cancel := context.WithCancel(context.Background())

	ram := &Ram{
		dao:    dao,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := ram.LoadSampleCache(); err != nil {
		return nil, err
	}
	return ram, nil
}

func (o *Ram) ShutdownSyncDB() {
	o.cancel()
	o.wg.Wait()
	logger.Println("[RAM] SyncDB goroutine is terminated")
}

func (o *Ram) StartSyncDB() error {
	o.wg.Add(1)
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		o.wg.Done()
	}()
	for {
		select {
		case <-o.ctx.Done():
			return nil
		case <-ticker.C:
			_, _, err := o.GetOprtModels(nil, true)
			if err != nil {
			}

			logger.Println("here sync DB")

			go func() {
				if err := o.dao.Ping(); err != nil {
					logger.Println("[RAM] Failed to sync DB")
				}
			}()
		}
	}
}
