#pragma once

#include <atomic>
#include <memory>
#include <thread>

namespace app {

class DbHandler;
class EventBus;
class RedisClient;

class SystemMonitoring {
public:
    SystemMonitoring(std::shared_ptr<DbHandler> db,
                     std::shared_ptr<RedisClient> redis,
                     std::shared_ptr<EventBus> bus);

    void start();
    void stop();

private:
    void run();

    std::shared_ptr<DbHandler>   db_;
    std::shared_ptr<RedisClient> redis_;
    std::shared_ptr<EventBus>    bus_;
    std::thread                  thread_;
    std::atomic<bool>            running_{false};
};

} // namespace app
