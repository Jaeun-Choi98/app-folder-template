package ram

import (
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/object"
	"sync"
)

type Ram struct {
	mu          sync.RWMutex
	dao         dbhandler.DBHandlerInterface
	SampleCache map[int]*object.Sample
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
