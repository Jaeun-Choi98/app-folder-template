package container

import (
	"pjt/internal/config"
	"pjt/internal/logger"
	repository "pjt/internal/repository/dao"
	"pjt/internal/service"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"

	"github.com/gorilla/mux"
)

type Container struct {
	Config       *config.Configuration
	CustomLogger *logger.CustomLogger
	Dao          repository.DaoInterface
	Service      service.ServcieInterface
	Controller   *controller.Controller
	RESTServer   *rest.RESTServer
}

func NewContainer() (*Container, error) {
	logger, err := logger.NewCustomLogger("")
	if err != nil {
		return nil, err
	}

	config, err := config.NewConfiguration()
	if err != nil {
		return nil, err
	}

	dao, err := repository.NewMyDB(config)
	if err != nil {
		return nil, err
	}

	service := service.NewMyServcie(dao, config)
	controller := controller.NewController(mux.NewRouter(), service, config)
	rest := rest.NewRESTServer(*controller, config)
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:       config,
		CustomLogger: logger,
		Dao:          dao,
		Service:      service,
		Controller:   controller,
		RESTServer:   rest,
	}, nil
}
