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
	Dao    repository.DaoInterface
	Config *config.Configuration
}

func NewMyServcie(dao repository.DaoInterface, config *config.Configuration) ServcieInterface {
	return &MyService{
		Dao:    dao,
		Config: config,
	}
}

func (s *MyService) Test() string {
	return s.Dao.Test()
}

func (s *MyService) Close() error {
	return s.Dao.Close()
}
