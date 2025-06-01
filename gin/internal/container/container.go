package container

import (
	"pjt/internal/config"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	"pjt/internal/service"
	apiservice "pjt/internal/service/api-service"
	sseservice "pjt/internal/service/sse-service"
	"pjt/internal/transport/eventbus"
	rest "pjt/internal/transport/http-rest"
	"pjt/internal/transport/http-rest/controller"
	"pjt/internal/transport/tcp"
	"time"

	"github.com/gin-gonic/gin"
)

var container *Container

type Container struct {
	Config     *config.Configuration
	Dao        dbhandler.DBHandlerInterface
	ApiService service.APIServcieInterface
	SseService service.SSEServiceInterface
	Controller *controller.Controller
	RESTServer *rest.RESTServer
	TCPServer  *tcp.TCPServer
	EventBus   *eventbus.EventBus
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

	dao, err := dbhandler.NewMyDB(config)
	if err != nil {
		return nil, err
	}

	ram, err := ram.NewRam(dao)
	if err != nil {
		return nil, err
	}

	eventbus := eventbus.NewEventBus()

	apiService := apiservice.NewAPIService(dao, ram, config)
	sseService := sseservice.NewSSEService()
	controller := controller.NewController(gin.New(), apiService, sseService, eventbus, config)
	rest := rest.NewRESTServer(*controller, config)

	tcp, err := tcp.NewTCPServer(eventbus, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:     config,
		Dao:        dao,
		ApiService: apiService,
		SseService: sseService,
		Controller: controller,
		RESTServer: rest,
		TCPServer:  tcp,
		EventBus:   eventbus,
	}, nil
}

// 컨테이너 객체(싱글톤 객체)를 nil로 초기화.
// nil이 아닐 경우 NewContainer()를 호출할 시 이전의 컨테이너 객체를 반환.
func (c *Container) Close() {
	container = nil
}
