#include "internal/healthcheck/monitoring.h"
#include "internal/db/db_handler/db_handler.h"
#include "internal/redis/redis_client.h"
#include "internal/event/event_bus.h"
#include "internal/event/event_define.h"
#include "internal/logger/logger.h"

#include <chrono>

namespace app {

SystemMonitoring::SystemMonitoring(std::shared_ptr<DbHandler> db,
                                   std::shared_ptr<RedisClient> redis,
                                   std::shared_ptr<EventBus> bus)
    : db_(std::move(db))
    , redis_(std::move(redis))
    , bus_(std::move(bus)) {}

void SystemMonitoring::start() {
    running_ = true;
    thread_  = std::thread(&SystemMonitoring::run, this);
}

void SystemMonitoring::stop() {
    running_ = false;
    if (thread_.joinable()) thread_.join();
}

void SystemMonitoring::run() {
    while (running_) {
        // TODO: DB/Redis 연결 상태 확인 → EventBus로 system.state 발행
        std::this_thread::sleep_for(std::chrono::seconds(30));
    }
}

} // namespace app
