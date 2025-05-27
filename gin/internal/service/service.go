package service

import (
	"fmt"
	"pjt/internal/config"
	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	repository "pjt/internal/repository/dao"
)

type ServcieInterface interface {
	Test() string
	Close() error

	// sse service
	GetSSEManager() (*SSESessionManager, error)
}

type MyService struct {
	Dao        repository.DaoInterface
	Ram        *ram.Ram
	SSEManager *SSESessionManager
	Config     *config.Configuration
}

func NewMyServcie(dao repository.DaoInterface, ram *ram.Ram, config *config.Configuration) ServcieInterface {
	return &MyService{
		Dao:        dao,
		Ram:        ram,
		Config:     config,
		SSEManager: NewSessionManager(),
	}
}

func (s *MyService) Test() string {
	return s.Dao.Test()
}

func (s *MyService) Close() error {
	s.SSEManager.Shutdown()
	return s.Dao.Close()
}

func (s *MyService) GetSSEManager() (*SSESessionManager, error) {
	if s.SSEManager == nil {
		logger.Println("SSEManager is empty")
		return nil, fmt.Errorf("sse manager is empty")
	}
	return s.SSEManager, nil
}
