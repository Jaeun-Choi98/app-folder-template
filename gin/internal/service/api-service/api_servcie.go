package service

import (
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/service"
)

type APIService struct {
	Dao    dbhandler.DBHandlerInterface
	Ram    *ram.Ram
	Config *config.Configuration
}

func NewAPIService(dao dbhandler.DBHandlerInterface, ram *ram.Ram, config *config.Configuration) service.APIServcieInterface {
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
