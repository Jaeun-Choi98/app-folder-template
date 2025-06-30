package container

import (
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	"pjt/internal/service"
	apiservice "pjt/internal/service/api-service"
	sseservice "pjt/internal/service/sse-service"
	tcpservice "pjt/internal/service/tcp-service"
	"pjt/internal/transport/eventbus"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"
	"pjt/internal/transport/monitoring"
	tcp "pjt/internal/transport/tcp/server"
	"time"

	"github.com/gin-gonic/gin"
)

var container *Container

type Container struct {
	Config           *config.Configuration
	Dao              dbhandler.DBHandlerInterface
	ApiService       service.APIServcieInterface
	SseService       service.SSEServiceInterface
	TcpServcie       service.TCPServiceInterface
	Controller       *controller.Controller
	RESTServer       *rest.RESTServer
	TCPServer        *tcp.TCPServer
	EventBus         *eventbus.EventBus
	SystemMonitoring *monitoring.SystemMonitoring
}

func NewContainer() (*Container, error) {

	if container != nil {
		return container, nil
	}
	customLogger, err := logger.NewCustomLogger("")
	if err != nil {
		return nil, err
	}
	// 전역 로거 객체 주입
	logger.SetLogger(customLogger)

	config, err := config.NewConfiguration()
	if err != nil {
		return nil, err
	}

	dao, err := dbhandler.NewDBHandler(config)
	if err != nil {
		return nil, err
	}

	ram, err := ram.NewRam(dao)
	if err != nil {
		return nil, err
	}

	eventbus := eventbus.NewEventBus()

	apiService := apiservice.NewAPIService(dao, ram, eventbus, config)
	sseService := sseservice.NewSSEService()
	tcpService := tcpservice.NewTCPService(eventbus)
	controller := controller.NewController(gin.New(), apiService, sseService, eventbus, config)
	rest := rest.NewRESTServer(*controller, config)

	tcp, err := tcp.NewTCPServer(eventbus, tcpService, 5*time.Second)
	if err != nil {
		return nil, err
	}

	monitoring := monitoring.NewSystemMonitoring(dao, tcp, eventbus, 1*time.Second)

	return &Container{
		Config:           config,
		Dao:              dao,
		ApiService:       apiService,
		SseService:       sseService,
		TcpServcie:       tcpService,
		Controller:       controller,
		RESTServer:       rest,
		TCPServer:        tcp,
		EventBus:         eventbus,
		SystemMonitoring: monitoring,
	}, nil
}
