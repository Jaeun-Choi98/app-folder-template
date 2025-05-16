package app

import (
	"context"
	"pjt/internal/container"
	rest "pjt/internal/transport/http-rest"
	"sync"
)

type Application struct {
	container  *container.Container
	restServer *rest.RESTServer
	wg         sync.WaitGroup
}

func NewApplication(container *container.Container) *Application {

	return &Application{
		container:  container,
		wg:         sync.WaitGroup{},
		restServer: container.RESTServer,
	}
}

func (a *Application) Start() error {
	return a.restServer.Start()
}

func (a *Application) Shutdown(ctx context.Context) error {
	return a.restServer.Shutdown(ctx)
}
