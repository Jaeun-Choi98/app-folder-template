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

	c.Router.HandleFunc("/sse-connect/{user_id:[0-9]+}", c.HandleSSE).Methods("GET")
	c.Router.HandleFunc("/sse-send", c.SendSSEMessageAll).Methods("POST")
	// JWT 인증이 필요한 SSE 엔드포인트
	//router.HandleFunc("/api/v1/events", middleware.JWTAuth(c.HandleSSE)).Methods("GET")

	// 특정 사용자에게 이벤트를 전송하는 관리 엔드포인트
	// 이 엔드포인트는 내부 서비스만 접근할 수 있어야 합니다
	//adminRouter := router.PathPrefix("/admin").Subrouter()
	//adminRouter.HandleFunc("/send-event/{userID}", middleware.AdminAuth(c.SendEventToUser)).Methods("POST")
	c.Router.Use(middleware.NewCORSMiddleware([]string{"*"}))
	c.Router.Use(middleware.LogMiddleware)
}

func (c *Controller) Close() error {
	return c.Service.Close()
}
