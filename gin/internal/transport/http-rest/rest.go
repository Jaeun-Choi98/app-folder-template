package rest

import (
	"context"
	"net/http"
	"pjt/internal/config"
	"pjt/internal/logger"
	"pjt/internal/transport/http-rest/controller"
	"time"
)

type RESTServer struct {
	server *http.Server
}

func NewRESTServer(controller controller.Controller, config *config.Configuration) *RESTServer {
	server := &http.Server{
		Addr: config.RestIp + ":" + config.RestPort,
		// if using sse, comment below
		// WriteTimeout: time.Second * 5,
		ReadTimeout: time.Second * 5,
		Handler:     controller.Router,
	}

	return &RESTServer{
		server: server,
	}
}

func (r *RESTServer) Start() error {
	return r.server.ListenAndServe()
}

func (r *RESTServer) Shutdown(ctx context.Context) error {
	logger.Println("[REST] REST goroutine terminated")
	return r.server.Shutdown(ctx)
}
