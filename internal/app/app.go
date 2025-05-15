package app

import (
	"context"
	"pjt/internal/container"
	repository "pjt/internal/repository/dao"
	"pjt/internal/service"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"
	"sync"
)

type Application struct {
	container  *container.Container
	restServer *rest.RESTServer
	wg         sync.WaitGroup
}

func NewApplication(container *container.Container) *Application {
	dao := repository.NewMyDB(container.Config)
	service := service.NewMyServcie(dao, container.Config)
	controller := controller.NewController(service, container.Config)
	rest := rest.NewRESTServer(*controller, container.Config)
	return &Application{
		container:  container,
		wg:         sync.WaitGroup{},
		restServer: rest,
	}
}

func (a *Application) Start() error {
	return a.restServer.Start()
}

func (a *Application) Shutdown(ctx context.Context) error {
	return a.restServer.Shutdown(ctx)
}
