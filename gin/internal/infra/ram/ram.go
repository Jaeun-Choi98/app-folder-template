package ram

import (
	"errors"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"sync"
)

/**
 * this layer is role of in memory db.
 * it needs reloading on start or modifing resource ( delete(both logical/physical), insert not update )
 *
 * when operating on get method. ram.cache -> service or other layer:
 * make new instance using mutex. it also convenient to set object
 * should serve pk filter, unique map
 * should except a logical deleted records.
 *
 */

var (
	errNotExistId = errors.New("not exists id ( some or one )")
)

type Ram struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int64]*object.Sample
}

func NewRam(dao dbhandler.DBHandlerInterface) (*Ram, error) {

	ram := &Ram{
		dao: dao,
	}

	if err := ram.LoadSampleCache(); err != nil {
		return nil, err
	}
	return ram, nil
}
