package controller

import (
	"net/http"
	"pjt/internal/config"
	"pjt/internal/service"
	"pjt/internal/transport/http-rest/middleware"

	"github.com/gorilla/mux"
)

type Controller struct {
	Router  *mux.Router
	Service service.ServcieInterface
	Config  *config.Configuration
}

func NewController(router *mux.Router, service service.ServcieInterface, config *config.Configuration) *Controller {

	controller := &Controller{
		Router:  router,
		Service: service,
		Config:  config,
	}
	controller.RoutePath()
	return controller
}

func (c *Controller) RoutePath() {
	c.Router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		str := c.Service.Test()
		w.Write([]byte(str))
	})
	c.Router.Use(middleware.NewCORSMiddleware([]string{"*"}))
	c.Router.Use(middleware.LogMiddleware)
}

func (c *Controller) Close() error {
	return c.Service.Close()
}
