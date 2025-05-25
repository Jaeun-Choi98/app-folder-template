package controller

import (
	"net/http"
	"pjt/internal/config"
	"pjt/internal/service"
	"pjt/internal/transport/http-rest/middleware"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	Router  *gin.Engine
	Service service.ServcieInterface
	Config  *config.Configuration
}

func NewController(router *gin.Engine, service service.ServcieInterface, config *config.Configuration) *Controller {

	controller := &Controller{
		Router:  router,
		Service: service,
		Config:  config,
	}
	controller.RoutePath()
	return controller
}

func (ctr *Controller) RoutePath() {
	ctr.Router.Use(middleware.LogMiddleware())
	ctr.Router.Use(middleware.NewCORSMiddleware([]string{"*"}))
	ctr.Router.GET("/test", func(c *gin.Context) {
		str := ctr.Service.Test()
		c.String(http.StatusOK, "%s", str)
	})

	// JWT 인증이 필요한 SSE 엔드포인트
	//router.HandleFunc("/api/v1/events", middleware.JWTAuth(c.HandleSSE)).Methods("GET")

	// 특정 사용자에게 이벤트를 전송하는 관리 엔드포인트
	// 이 엔드포인트는 내부 서비스만 접근할 수 있어야 합니다
	//adminRouter := router.PathPrefix("/admin").Subrouter()
	//adminRouter.HandleFunc("/send-event/{userID}", middleware.AdminAuth(c.SendEventToUser)).Methods("POST")

}

func (ctr *Controller) Close() error {
	return ctr.Service.Close()
}
