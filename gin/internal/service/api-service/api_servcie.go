package service

import (
	"pjt/internal/config"
	"pjt/internal/infra/ram"
	repository "pjt/internal/repository/dao"
	"pjt/internal/service"
)

type APIService struct {
	Dao    repository.DaoInterface
	Ram    *ram.Ram
	Config *config.Configuration
}

func NewAPIService(dao repository.DaoInterface, ram *ram.Ram, config *config.Configuration) service.APIServcieInterface {
	return &APIService{
		Dao:    dao,
		Ram:    ram,
		Config: config,
	}
}

func (a *APIService) Test() string {
	return a.Dao.Test()
}

func (a *APIService) Close() error {
	return a.Dao.Close()
}
