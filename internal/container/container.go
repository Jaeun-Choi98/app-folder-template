package container

import "pjt/internal/config"

type Container struct {
	Config *config.Configuration
}

func NewContainer() (*Container, error) {

	config, err := config.NewConfiguration()
	if err != nil {
		return nil, err
	}

	return &Container{
		Config: config,
	}, nil
}
