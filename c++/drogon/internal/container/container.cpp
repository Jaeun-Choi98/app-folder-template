#include "internal/container/container.h"

#include "internal/db/db_handler/db_handler.h"
#include "internal/event/event_bus.h"
#include "internal/healthcheck/monitoring.h"
#include "internal/infra/cache/cache.h"
#include "internal/logger/logger.h"
#include "internal/redis/redis_client.h"
#include "internal/service/api_service/api_service.h"
#include "internal/service/tcp_service/tcp_service.h"
#include "internal/transport/http_rest/http_utils/jwt_util.h"
#include "internal/transport/http_rest/rest.h"
#include "internal/transport/tcp/server/client/client_manager.h"
#include "internal/transport/tcp/server/tcp_server.h"
#include "internal/worker/cache_worker.h"
#include "internal/worker/cron_worker.h"

namespace app {

Container Container::build(const std::string& configPath) {
  Container c;
  c.config = Config::load(configPath);

  // Logger
  Logger::instance().init("log");

  // 인프라
  c.eventBus = std::make_shared<EventBus>();
  c.cache = std::make_shared<Cache>();
  c.redisClient = std::make_shared<RedisClient>();

  // DB
  c.dbHandler = std::shared_ptr<DbHandler>(DbHandler::create(c.config));
  c.dbHandler->connect();

  // Redis
  if (!c.config.redisHost.empty()) {
    c.redisClient->connect(c.config);
  }

  // JWT
  c.jwtUtil =
      std::make_shared<JwtUtil>(c.config.jwtSecret, c.config.jwtExpireMin);

  // 서비스
  c.apiService = std::make_shared<ApiServiceImpl>(c.dbHandler, c.cache,
                                                  c.eventBus, c.redisClient);
  c.tcpService = std::make_shared<TcpServiceImpl>(c.dbHandler, c.cache);

  // 트랜스포트
  c.clientManager = std::make_shared<ClientManager>();
  c.restServer = std::make_shared<RestServer>(c.config, c.apiService,
                                              c.eventBus, c.jwtUtil);
  c.tcpServer =
      std::make_shared<TcpServer>(c.config, c.clientManager, c.tcpService);

  // 워커
  c.cacheWorker =
      std::make_shared<CacheWorker>(c.cache, c.dbHandler, c.eventBus);
  c.cronWorker = std::make_shared<CronWorker>();
  c.monitoring = std::make_shared<SystemMonitoring>(c.dbHandler, c.redisClient,
                                                    c.eventBus);

  return c;
}

}  // namespace app
