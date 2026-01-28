package container

import (
	"context"
	"pjt/internal/config"

	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/redis"
	"pjt/internal/worker"

	"pjt/internal/infra/cache"
	"pjt/internal/logger"
	eventlog "pjt/internal/logger/event-logger"
	"pjt/internal/service"
	apiservice "pjt/internal/service/api-service"

	tcpservice "pjt/internal/service/tcp-service"

	monitoring "pjt/internal/healthcheck"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"
	tcp "pjt/internal/transport/tcp/server"
	"pjt/internal/transport/tcp/server/client"
	"time"

	"github.com/Jaeun-Choi98/modules/eventbus"
	"github.com/Jaeun-Choi98/modules/sse"
	"github.com/gin-gonic/gin"
)

var container *Container

type Container struct {
	Config           *config.Configuration
	Dao              dbhandler.DBHandlerInterface
	Cache            *cache.Cache
	ApiService       service.APIServcieInterface
	SseManager       *sse.SessionManager[uint32, uint32]
	TcpService       service.TCPServiceInterface
	Controller       *controller.Controller
	RESTServer       *rest.RESTServer
	TCPServer        *tcp.TCPServer
	EventBus         *eventbus.EventBus
	SystemMonitoring *monitoring.SystemMonitoring

	CacheWorker *worker.CacheWorker
	TCPWorker   *worker.TCPWorker
	CronWorker  *worker.CronWorker
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

	cache, err := cache.NewCacheMem(dao)
	if err != nil {
		return nil, err
	}

	eventbus := eventbus.NewEventBus(context.Background(), 30*time.Minute)

	apiService := apiservice.NewAPIService(dao, cache, eventbus, config)
	sseManager := sse.NewSessionManager[uint32, uint32]()
	controller := controller.NewController(gin.New(), apiService, sseManager, eventbus, config)
	rest := rest.NewRESTServer(*controller, config)

	tcpClientManager := client.NewClientManager()
	tcpService := tcpservice.NewTCPService(tcpClientManager, dao, cache, eventbus)
	tcp, err := tcp.NewTCPServer(tcpClientManager, config, tcpService, 5*time.Second, 5)
	if err != nil {
		return nil, err
	}

	cacheWorker := worker.NewCacheWorker(dao, cache)
	tcpWorker := worker.NewTCPWorker(dao, cache, eventbus, tcpClientManager)
	cronWorker := worker.NewCronWorker(config, dao, eventbus)

	monitoring := monitoring.NewSystemMonitoring(dao, tcp, eventbus, 1*time.Second)

	dbLogger, _ := eventlog.NewDBLogger(dao, cache, eventbus)
	eventlog.SetEventLogger(dbLogger)

	if err := redis.InitRedis(config); err != nil {
		return nil, err
	}

	return &Container{
		Config:           config,
		Dao:              dao,
		Cache:            cache,
		ApiService:       apiService,
		SseManager:       sseManager,
		TcpService:       tcpService,
		Controller:       controller,
		RESTServer:       rest,
		TCPServer:        tcp,
		EventBus:         eventbus,
		CacheWorker:      cacheWorker,
		TCPWorker:        tcpWorker,
		CronWorker:       cronWorker,
		SystemMonitoring: monitoring,
	}, nil
}
