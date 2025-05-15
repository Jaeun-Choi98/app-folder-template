package service

import (
	"pjt/root/internal/config"
	repository "pjt/root/internal/repository/dao"
)

type ServcieInterface interface {
	Test() string
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
