#pragma once

#include "internal/config/config.h"

#include <memory>

namespace app {

class DbHandler;
class Cache;
class EventBus;
class RedisClient;
class ApiServiceImpl;
class TcpServiceImpl;
class RestServer;
class TcpServer;
class ClientManager;
class CacheWorker;
class CronWorker;
class SystemMonitoring;
class JwtUtil;

struct Container {
    Config config;

    std::shared_ptr<DbHandler>        dbHandler;
    std::shared_ptr<Cache>            cache;
    std::shared_ptr<EventBus>         eventBus;
    std::shared_ptr<RedisClient>      redisClient;
    std::shared_ptr<JwtUtil>          jwtUtil;

    std::shared_ptr<ApiServiceImpl>   apiService;
    std::shared_ptr<TcpServiceImpl>   tcpService;

    std::shared_ptr<ClientManager>    clientManager;
    std::shared_ptr<RestServer>       restServer;
    std::shared_ptr<TcpServer>        tcpServer;

    std::shared_ptr<CacheWorker>      cacheWorker;
    std::shared_ptr<CronWorker>       cronWorker;
    std::shared_ptr<SystemMonitoring> monitoring;

    static Container build(const std::string& configPath);
};

} // namespace app
