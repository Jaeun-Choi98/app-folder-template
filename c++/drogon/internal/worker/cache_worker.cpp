#include "internal/worker/cache_worker.h"
#include "internal/infra/cache/cache.h"
#include "internal/infra/define/define.h"
#include "internal/db/db_handler/db_handler.h"
#include "internal/event/event_bus.h"
#include "internal/logger/logger.h"

#include <chrono>

namespace app {

CacheWorker::CacheWorker(std::shared_ptr<Cache> cache,
                         std::shared_ptr<DbHandler> db,
                         std::shared_ptr<EventBus> bus)
    : cache_(std::move(cache))
    , db_(std::move(db))
    , bus_(std::move(bus)) {}

void CacheWorker::start() {
    running_ = true;
    thread_  = std::thread(&CacheWorker::run, this);
}

void CacheWorker::stop() {
    running_ = false;
    if (thread_.joinable()) thread_.join();
}

void CacheWorker::run() {
    while (running_) {
        std::this_thread::sleep_for(
            std::chrono::milliseconds(kCacheWorkerIntervalMs));

        if (!cache_->isDirty()) continue;

        // TODO: 변경된 캐시 데이터를 DB에 동기화
        cache_->resetDirty();
    }
}

} // namespace app
