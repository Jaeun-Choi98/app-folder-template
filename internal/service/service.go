package service

import (
	"pjt/internal/config"
	repository "pjt/internal/repository/dao"
)

type ServcieInterface interface {
	Test() string
	Close() error
}

type MyService struct {
	Dao        repository.DaoInterface
	SSEManager *SSESessionManager
	Config     *config.Configuration
}

func NewMyServcie(dao repository.DaoInterface, config *config.Configuration) ServcieInterface {
	return &MyService{
		Dao:        dao,
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
