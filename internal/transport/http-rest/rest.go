package rest

import (
	"context"
	"net/http"
	"pjt/root/internal/config"
	"pjt/root/internal/transport/http-rest/controller"
	"time"
)

type RESTServer struct {
	server *http.Server
}

func NewRESTServer(controller controller.Controller, config *config.Configuration) *RESTServer {
	server := &http.Server{
		Addr:         config.RestIp + ":" + config.RestPort,
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		Handler:      controller.Router,
	}

	return &RESTServer{
		server: server,
	}
}

func (r *RESTServer) Start() error {
	return r.server.ListenAndServe()
}

func (r *RESTServer) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
