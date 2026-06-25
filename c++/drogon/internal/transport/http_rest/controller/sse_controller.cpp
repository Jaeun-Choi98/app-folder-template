#include "internal/transport/http_rest/controller/sse_controller.h"
#include "internal/event/event_bus.h"
#include "internal/event/event_define.h"

namespace app {

void SseController::setEventBus(std::shared_ptr<EventBus> bus) {
    bus_ = std::move(bus);
}

void SseController::stream(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback) {
    // TODO: Drogon의 streaming response를 사용하여 SSE 구현
    // 1. Content-Type: text/event-stream 설정
    // 2. EventBus에서 cache.data / system.state / system.event 토픽 구독
    // 3. 이벤트 수신 시 "data: ...\n\n" 형식으로 전송

    auto resp = drogon::HttpResponse::newHttpResponse();
    resp->setContentTypeCode(drogon::CT_NONE);
    resp->addHeader("Content-Type", "text/event-stream");
    resp->addHeader("Cache-Control", "no-cache");
    resp->addHeader("Connection", "keep-alive");
    resp->addHeader("Access-Control-Allow-Origin", "*");
    callback(resp);
}

} // namespace app
