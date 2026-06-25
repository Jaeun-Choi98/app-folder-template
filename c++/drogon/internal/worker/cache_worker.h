#pragma once

#include <atomic>
#include <memory>
#include <thread>

namespace app {

class Cache;
class DbHandler;
class EventBus;

class CacheWorker {
public:
    CacheWorker(std::shared_ptr<Cache> cache,
                std::shared_ptr<DbHandler> db,
                std::shared_ptr<EventBus> bus);

    void start();
    void stop();

private:
    void run();

    std::shared_ptr<Cache>     cache_;
    std::shared_ptr<DbHandler> db_;
    std::shared_ptr<EventBus>  bus_;
    std::thread                thread_;
    std::atomic<bool>          running_{false};
};

} // namespace app
