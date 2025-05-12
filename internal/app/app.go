package app

import (
	"context"
	"root/internal/container"
	"sync"
)

type Application struct {
	container *container.Container
	wg        sync.WaitGroup
}

func NewApplication(container *container.Container) *Application {
	return &Application{
		container: container,
		wg:        sync.WaitGroup{},
	}
}

func (a *Application) Start() error {
	return nil
}

func (a *Application) Shutdown(ctx context.Context) error {
	return nil
}
