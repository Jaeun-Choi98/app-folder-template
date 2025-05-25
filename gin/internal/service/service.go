package service

import (
	"pjt/internal/config"
	"pjt/internal/infra/ram"
	repository "pjt/internal/repository/dao"
)

type ServcieInterface interface {
	Test() string
	Close() error

	// sse service

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
