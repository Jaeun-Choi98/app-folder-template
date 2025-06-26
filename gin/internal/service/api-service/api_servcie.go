package service

import (
	"errors"
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/service"
	"pjt/internal/transport/eventbus"
	"sync"
)

/**
 * in service layer, varifing reference integrity and unique key.
 * also, doing default value injection about patch api.
 * e.g.
 * delete -> if parent table, verify referenced key
 * update, insert -> if child table, verify reference key, and verify unique key ( logical or physical )
 *
 *
 * service layer is role of view about db or ram.cache(im memory db)
 * it means processing pk and logical delete(column:DelYN) in db or ram.cache
 *
 * also, doing processing nil value about patch api. ( default value: using ram.cache or db )
 */

var (
	errExistsNum  = errors.New("num already exists")
	errExistsID   = errors.New("ID already exists")
	errExistsName = errors.New("name already exists")

	// 아래 에러는 식별자가 아닌 값으로 검사를 진행할 때 사용
	errNotExistsName = errors.New("not exists name")

	errRefIntegrity = errors.New("referential integrity is not satisfied")

	errInvalidID = errors.New("invalid ID")
	errInvalidPW = errors.New("invalid PW")
)

type APIService struct {
	Dao      dbhandler.DBHandlerInterface
	Ram      *ram.Ram
	Config   *config.Configuration
	EventBus *eventbus.EventBus
}

func NewAPIService(dao dbhandler.DBHandlerInterface, ram *ram.Ram, eventBus *eventbus.EventBus, config *config.Configuration) service.APIServcieInterface {
	return &APIService{
		Dao:      dao,
		Ram:      ram,
		Config:   config,
		EventBus: eventBus,
	}
}

func (a *APIService) Test() string {
	return a.Dao.Test()
}

func (a *APIService) Close() error {
	return a.Dao.Close()
}

// 로그인 세션
type LoginSeesion struct {
	Acc map[int64]*string // MemberId -> accToken
	Ref map[int64]*string // MemberId -> refToken
	mu  sync.RWMutex
}

func NewLoginSession() *LoginSeesion {
	return &LoginSeesion{
		Acc: make(map[int64]*string),
		Ref: make(map[int64]*string),
	}
}

func (l *LoginSeesion) SetAcc(id int64, token *string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Acc[id] = token
}

func (l *LoginSeesion) SetRef(id int64, token *string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Ref[id] = token
}

func (l *LoginSeesion) GetAcc(id int64) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if token, exists := l.Acc[id]; exists {
		return *token
	}
	return ""
}

func (l *LoginSeesion) GetRef(id int64) string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if token, exists := l.Ref[id]; exists {
		return *token
	}
	return ""
}

func (l *LoginSeesion) DeleteAcc(id int64, token string, auto bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, exists := l.Acc[id]; exists {
		if auto {
			if *v == token {
				delete(l.Acc, id)
				return true
			} else {
				return false
			}
		} else {
			delete(l.Acc, id)
			return true
		}
	} else {
		return false
	}
}

func (l *LoginSeesion) DeleteRef(id int64, token string, auto bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if v, exists := l.Ref[id]; exists {
		if auto {
			if *v == token {
				delete(l.Ref, id)
				return true
			} else {
				return false
			}
		} else {
			delete(l.Ref, id)
			return true
		}
	} else {
		return false
	}
}

func (l *LoginSeesion) IsExistsAcc(id int64) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, exists := l.Acc[id]
	return exists
}

func (l *LoginSeesion) IsExistsRef(id int64) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, exists := l.Ref[id]
	return exists
}
