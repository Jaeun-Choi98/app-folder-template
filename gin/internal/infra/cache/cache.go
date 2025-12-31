package cache

import (
	"context"
	"errors"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"pjt/internal/logger"
	"sync"
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

type Cache struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int64]*object.Sample
}

func NewCacheMem(dao dbhandler.DBHandlerInterface) (*Cache, error) {

	ctx, cancel := context.WithCancel(context.Background())

	ram := &Cache{
		dao:    dao,
		ctx:    ctx,
		cancel: cancel,
	}

	if err := ram.LoadSampleCache(); err != nil {
		return nil, err
	}
	return ram, nil
}

func (o *Cache) Close() {
	o.cancel()
	o.wg.Wait()
	logger.Println("[Cache] SyncDB goroutine is terminated")
}
