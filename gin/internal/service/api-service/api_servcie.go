package service

import (
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/service"
	"pjt/internal/transport/eventbus"
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
