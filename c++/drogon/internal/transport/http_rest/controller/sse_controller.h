#pragma once

#include <drogon/HttpController.h>
#include <memory>

namespace app {

class EventBus;

class SseController : public drogon::HttpController<SseController> {
public:
    METHOD_LIST_BEGIN
    ADD_METHOD_TO(SseController::stream, "/api/sse/events", drogon::Get);
    METHOD_LIST_END

    void stream(const drogon::HttpRequestPtr& req,
                std::function<void(const drogon::HttpResponsePtr&)>&& callback);

    void setEventBus(std::shared_ptr<EventBus> bus);

private:
    std::shared_ptr<EventBus> bus_;
};

} // namespace app
