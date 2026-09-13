package controller

import (
	"errors"
	"net/http"
	"pjt/internal/config"
	"pjt/internal/event"
	"pjt/internal/logger"
	"pjt/internal/service"
	"strconv"

	"pjt/internal/transport/http-rest/http-utils/httperr"
	"pjt/internal/transport/http-rest/http-utils/jwt"
	"pjt/internal/transport/http-rest/middleware"
	"pjt/internal/transport/http-rest/response"
	"time"

	"github.com/Jaeun-Choi98/modules/eventbus"
	"github.com/Jaeun-Choi98/modules/sse"
	"github.com/Jaeun-Choi98/modules/utils"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	Router     *gin.Engine
	ApiService service.APIServcieInterface
	SseManager *sse.SessionManager[uint32, uint32]
	Config     *config.Configuration
}

func NewController(router *gin.Engine, apiService service.APIServcieInterface, sseManager *sse.SessionManager[uint32, uint32], config *config.Configuration) *Controller {

	controller := &Controller{
		Router:     router,
		ApiService: apiService,
		SseManager: sseManager,
		Config:     config,
	}
	controller.RoutePath()
	return controller
}

func (c *Controller) RoutePath() {
	c.Router.Use(middleware.LogMiddleware())
	c.Router.Use(middleware.ErrorMiddleware())
	c.Router.Use(middleware.NewCORSMiddleware(c.Config.Cors, true))

	// 로그 레벨 / 덤프 여부 조회 및 실행 중 변경 ( 재기동 시 env.ini 값으로 복귀 )
	// ex) curl 'host/log-config?level=debug&dump=false'
	c.Router.GET("/log-config", c.HandleGetLogConfig)

	// 쿠키를 사용해서 jwt 토큰을 전달
	c.Router.GET("/test", func(ctx *gin.Context) {
		str := c.ApiService.Test()
		jwt, err := jwt.NewJwtHS256(1, "name")
		if err != nil {
			ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
			return
		}
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:    "jwt",
			Value:   jwt,
			MaxAge:  60 * 60,
			Expires: time.Now().Add(time.Minute * 60).In(utils.LocalKorea),
			Path:    "/",
		})
		ctx.String(http.StatusOK, "%s", str)
	})

	needJwt := c.Router.Group("/jwt")
	needJwt.Use(middleware.JWTMiddleware())
	needJwt.GET("/test", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "%s", "success jwt test")
	})
	needJwt.GET("/remove-test", func(ctx *gin.Context) {
		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:    "jwt",
			Value:   "",
			MaxAge:  60 * 60,
			Expires: time.Now().Add(time.Minute * 60).In(utils.LocalKorea),
			Path:    "/",
		})
		ctx.String(http.StatusOK, "%s", "remove-test ok")
	})

	c.Router.POST("/sse-send", c.SendSSEMessageAll)
	/**
	 * 특정 사용자에게 이벤트를 전송하는 관리 엔드포인트
	 * 이 엔드포인트는 내부 서비스만 접근할 수 있어야 합니다
	 */
	c.Router.POST("/sse-send/:id", c.SendSSEMessageToUser)

	sseConnect := c.Router.Group("/sse")
	sseConnect.Use(middleware.SSEMiddleware(), middleware.StoreIdToContext())
	sseConnect.GET("/connect", c.HandleSSEConnect)

	// TCP 송신 테스트, 테스트 하려면 tcp/server.go 에서 클라이언트 ID를 임의로 설정해야함.
	c.Router.GET("/send-work", func(ctx *gin.Context) {
		if err := eventbus.PublishSync(event.GetEventBus(), "tcp.event", &event.TcpEvent{
			TcpEventOpcode: 0x01,
			ClientId:       1,
			PacketOpcode:   0x02,
			Data:           []byte("no reply"),
			SendTimeout:    5 * time.Second,
		}); err != nil {
			ctx.Error(httperr.INNER_ERROR.Add(err[0], response.FAIL))
			return
		}

		ctx.JSON(http.StatusOK, response.SUCCESS)
	})

	c.Router.GET("/send-work-reply", func(ctx *gin.Context) {
		results, errs := eventbus.Request(event.GetEventBus(), "tcp.event", &event.TcpEvent{
			TcpEventOpcode: 0x01,
			ClientId:       1,
			PacketOpcode:   0x02,
			Data:           []byte("no reply"),
			SendTimeout:    5 * time.Second,
			ReplyTimeout:   5 * time.Second,
		}, 8*time.Second)

		if errs != nil {
			ctx.Error(httperr.INNER_ERROR.Add(errs[0], response.FAIL))
			return
		}

		ctx.JSON(http.StatusOK, response.SUCCESS.Add(results))
	})

	// 웹서버
	// c.Router.GET("/", func(c *gin.Context) { c.File(filepath.Join("build", "index.html")) })
	c.Router.Use(middleware.SpaHandlerOther("/spatest", "spa-test"))
	c.Router.Use(middleware.SpaHandlerRoot("build", "index.html"))
}

func (ctr *Controller) Close() error {
	return ctr.ApiService.Close()
}

// HandleGetLogConfig 는 현재 로그 레벨/덤프 여부를 돌려준다.
// 쿼리를 주면 그 값만 바꾼 뒤 돌려준다 (curl 로 쓰기 쉽게 GET 으로 둔다).
//
//	GET /log-config                      조회
//	GET /log-config?level=debug          레벨 변경 (debug/info/warn/error)
//	GET /log-config?dump=false           덤프 끄기
//
// 잘못된 값이 하나라도 있으면 아무것도 바꾸지 않는다.
func (t *Controller) HandleGetLogConfig(c *gin.Context) {
	levelStr, hasLevel := c.GetQuery("level")
	dumpStr, hasDump := c.GetQuery("dump")

	var level logger.Level
	var dump bool
	var err error
	if hasLevel {
		if level, err = logger.ParseLevel(levelStr); err != nil {
			c.Error(httperr.BADREQUEST.Add(err, response.INVAILD_DATA))
			return
		}
	}
	if hasDump {
		if dump, err = strconv.ParseBool(dumpStr); err != nil {
			c.Error(httperr.BADREQUEST.Add(errors.New("invalid dump value: "+dumpStr), response.INVAILD_DATA))
			return
		}
	}

	if hasLevel {
		logger.SetLevel(level)
	}
	if hasDump {
		logger.SetDump(dump)
	}

	cur := gin.H{"level": logger.GetLevel().String(), "dump": logger.IsDumpEnabled()}
	if hasLevel || hasDump {
		logger.Warnf("[REST] log config changed: level=%v, dump=%v, remote: %s", cur["level"], cur["dump"], c.ClientIP())
	}
	c.JSON(http.StatusOK, response.SUCCESS.Add(cur))
}
