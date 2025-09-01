package container

import (
	"context"
	"pjt/internal/config"
	"pjt/internal/cron"
	dbhandler "pjt/internal/db/db-handler"

	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	eventlog "pjt/internal/logger/event-logger"
	"pjt/internal/service"
	apiservice "pjt/internal/service/api-service"
	sseservice "pjt/internal/service/sse-service"
	tcpservice "pjt/internal/service/tcp-service"
	"pjt/internal/transport/eventbus"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"
	"pjt/internal/transport/monitoring"
	tcp "pjt/internal/transport/tcp/server"
	"pjt/internal/transport/tcp/server/client"
	"time"

	"github.com/gin-gonic/gin"
)

var container *Container

type Container struct {
	Config           *config.Configuration
	Dao              dbhandler.DBHandlerInterface
	Ram              *ram.Ram
	ApiService       service.APIServcieInterface
	SseService       service.SSEServiceInterface
	TcpService       service.TCPServiceInterface
	Controller       *controller.Controller
	RESTServer       *rest.RESTServer
	TCPServer        *tcp.TCPServer
	EventBus         *eventbus.EventBus
	Cron             *cron.Cron
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

	eventbus := eventbus.NewEventBus(context.Background(), 30*time.Minute)

	apiService := apiservice.NewAPIService(dao, ram, eventbus, config)
	sseService := sseservice.NewSSEService()
	controller := controller.NewController(gin.New(), apiService, sseService, eventbus, config)
	rest := rest.NewRESTServer(*controller, config)

	tcpClientManager := client.NewClientManager()
	tcpService := tcpservice.NewTCPService(tcpClientManager, dao, ram, eventbus)
	tcp, err := tcp.NewTCPServer(tcpClientManager, tcpService, 5*time.Second)
	if err != nil {
		return nil, err
	}

	monitoring := monitoring.NewSystemMonitoring(dao, tcp, eventbus, 1*time.Second)

	dbLogger, _ := eventlog.NewDBLogger(dao, ram, eventbus)
	eventlog.SetEventLogger(dbLogger)
	cron := cron.NewCron(config, dao, eventbus)

	return &Container{
		Config:           config,
		Dao:              dao,
		Ram:              ram,
		ApiService:       apiService,
		SseService:       sseService,
		TcpService:       tcpService,
		Controller:       controller,
		RESTServer:       rest,
		TCPServer:        tcp,
		EventBus:         eventbus,
		Cron:             cron,
		SystemMonitoring: monitoring,
	}, nil
}
