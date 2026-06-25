#include "internal/transport/http_rest/rest.h"
#include "internal/service/service.h"
#include "internal/event/event_bus.h"
#include "internal/transport/http_rest/http_utils/jwt_util.h"

#include <drogon/drogon.h>

namespace app {

RestServer::RestServer(const Config& cfg,
                       std::shared_ptr<ApiService> svc,
                       std::shared_ptr<EventBus> bus,
                       std::shared_ptr<JwtUtil> jwt)
    : cfg_(cfg)
    , svc_(std::move(svc))
    , bus_(std::move(bus))
    , jwt_(std::move(jwt)) {}

void RestServer::start() {
    thread_ = std::thread([this] {
        drogon::app()
            .setLogPath("")
            .setLogLevel(trantor::Logger::kWarn)
            .addListener(cfg_.serverHost, cfg_.serverPort)
            .setThreadNum(4)
            .run();
    });
}

void RestServer::stop() {
    drogon::app().quit();
    if (thread_.joinable()) thread_.join();
}

} // namespace app
