package ram

import (
	model "pjt/internal/model/sample"
	repository "pjt/internal/repository/dao"
	"sync"
)

type Ram struct {
	mu         *sync.RWMutex
	dao        repository.DaoInterface
	OprtModels map[int]*model.SampleModel
}

func NewRam(dao repository.DaoInterface) (*Ram, error) {

	ram := &Ram{
		mu:  &sync.RWMutex{},
		dao: dao,
	}

	if err := ram.LoadOprtModels(); err != nil {
		return nil, err
	}
	return ram, nil
}

func (r *Ram) LoadOprtModels() error {

	r.mu.Lock()
	defer r.mu.Unlock()

	return nil
}

// 공유 자원 동기화를 위한 새로운 메모리 할당
func (r *Ram) GetOprtModels() map[int]model.SampleModel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return nil
}
