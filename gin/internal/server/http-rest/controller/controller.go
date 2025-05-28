package controller

import (
	"net/http"
	"pjt/internal/config"
	model "pjt/internal/model/sample"
	"pjt/internal/server/http-rest/http-utils/httperr"
	"pjt/internal/server/http-rest/http-utils/jwt"
	"pjt/internal/server/http-rest/middleware"
	"pjt/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	Router     *gin.Engine
	ApiService service.APIServcieInterface
	SseService service.SSEServiceInterface
	Config     *config.Configuration
}

func NewController(router *gin.Engine, apiService service.APIServcieInterface, sseService service.SSEServiceInterface, config *config.Configuration) *Controller {

	controller := &Controller{
		Router:     router,
		ApiService: apiService,
		SseService: sseService,
		Config:     config,
	}
	controller.RoutePath()
	return controller
}

func (c *Controller) RoutePath() {
	c.Router.Use(middleware.ErrorMiddleware())
	c.Router.Use(middleware.LogMiddleware())
	c.Router.Use(middleware.NewCORSMiddleware([]string{"*"}))

	// 쿠키를 사용해서 jwt 토큰을 전달
	c.Router.GET("/test", func(ctx *gin.Context) {
		str := c.ApiService.Test()
		jwt, err := jwt.NewJwtHS256(&model.SampleModel{Id: 1, Name: "cju"})
		if err != nil {
			ctx.Error(httperr.INNER_ERROR.AddErrMsg(err))
			return
		}
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:    "jwt",
			Value:   jwt,
			MaxAge:  60 * 60,
			Expires: time.Now().Add(time.Minute * 60),
			Path:    "/",
		})
		ctx.String(http.StatusOK, "%s", str)
	})

	needJwt := c.Router.Group("/jwt")
	needJwt.Use(middleware.JWTMiddleware())
	needJwt.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "%s", "success jwt test")
	})

	// JWT 인증이 필요한 SSE 엔드포인트
	//router.HandleFunc("/api/v1/events", middleware.JWTAuth(c.HandleSSE)).Methods("GET")

	// 특정 사용자에게 이벤트를 전송하는 관리 엔드포인트
	// 이 엔드포인트는 내부 서비스만 접근할 수 있어야 합니다
	//adminRouter := router.PathPrefix("/admin").Subrouter()
	//adminRouter.HandleFunc("/send-event/{userID}", middleware.AdminAuth(c.SendEventToUser)).Methods("POST")
	c.Router.GET("/sse-connect/:id", c.HandleSSEConnect)
	c.Router.POST("/sse-send", c.SendSSEMessageAll)
}

func (ctr *Controller) Close() error {
	return ctr.ApiService.Close()
}
