#include "internal/app/app.h"
#include "internal/logger/logger.h"
#include "internal/transport/http_rest/rest.h"
#include "internal/transport/tcp/server/tcp_server.h"
#include "internal/worker/cache_worker.h"
#include "internal/worker/cron_worker.h"
#include "internal/healthcheck/monitoring.h"
#include "internal/event/event_bus.h"
#include "internal/redis/redis_client.h"
#include "internal/db/db_handler/db_handler.h"

#include <csignal>
#include <condition_variable>
#include <mutex>

namespace {
    std::mutex              gMtx;
    std::condition_variable gCv;
    bool                    gShutdown = false;

    void signalHandler(int) {
        std::lock_guard lock(gMtx);
        gShutdown = true;
        gCv.notify_all();
    }
}

namespace app {

Application::Application(Container container)
    : c_(std::move(container)) {}

void Application::start() {
    Logger::instance().info("application starting...");

    c_.monitoring->start();
    c_.cacheWorker->start();
    c_.cronWorker->start();
    c_.tcpServer->start();
    c_.restServer->start();

    Logger::instance().info("application started");
}

void Application::shutdown() {
    Logger::instance().info("application shutting down...");

    // 종료 순서: Monitoring → REST → CacheWorker → CronWorker → TCP → EventBus → Redis → Logger
    c_.monitoring->stop();
    c_.restServer->stop();
    c_.cacheWorker->stop();
    c_.cronWorker->stop();
    c_.tcpServer->stop();
    c_.eventBus->clear();
    c_.redisClient->disconnect();
    c_.dbHandler->disconnect();

    Logger::instance().info("application stopped");
}

void Application::waitForSignal() {
    std::signal(SIGINT, signalHandler);
    std::signal(SIGTERM, signalHandler);

    std::unique_lock lock(gMtx);
    gCv.wait(lock, [] { return gShutdown; });
}

} // namespace app
