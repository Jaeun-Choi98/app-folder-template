#pragma once

#include "internal/config/config.h"

#include <memory>
#include <thread>

namespace app {

class ApiService;
class EventBus;
class JwtUtil;

class RestServer {
public:
    RestServer(const Config& cfg,
               std::shared_ptr<ApiService> svc,
               std::shared_ptr<EventBus> bus,
               std::shared_ptr<JwtUtil> jwt);

    void start();
    void stop();

private:
    Config                      cfg_;
    std::shared_ptr<ApiService> svc_;
    std::shared_ptr<EventBus>   bus_;
    std::shared_ptr<JwtUtil>    jwt_;
    std::thread                 thread_;
};

} // namespace app
