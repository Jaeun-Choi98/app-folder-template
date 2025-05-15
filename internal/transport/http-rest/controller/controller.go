package controller

import (
	"net/http"
	"pjt/root/internal/config"
	"pjt/root/internal/service"
	"pjt/root/internal/transport/http-rest/middleware"

	"github.com/gorilla/mux"
)

type Controller struct {
	Router  *mux.Router
	Service service.ServcieInterface
	Config  *config.Configuration
}

func NewController(service service.ServcieInterface, config *config.Configuration) *Controller {
	r := mux.NewRouter()
	controller := &Controller{
		Router:  r,
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
}
