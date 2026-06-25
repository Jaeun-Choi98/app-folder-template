#pragma once

#include <drogon/HttpController.h>
#include <memory>

namespace app {

class ApiService;
class EventBus;

class ApiController : public drogon::HttpController<ApiController> {
public:
    METHOD_LIST_BEGIN
    ADD_METHOD_TO(ApiController::getAll,      "/api/samples",     drogon::Get);
    ADD_METHOD_TO(ApiController::getById,     "/api/samples/{id}", drogon::Get);
    ADD_METHOD_TO(ApiController::create,      "/api/samples",     drogon::Post);
    ADD_METHOD_TO(ApiController::updateById,  "/api/samples/{id}", drogon::Put);
    ADD_METHOD_TO(ApiController::deleteById,  "/api/samples/{id}", drogon::Delete);
    ADD_METHOD_TO(ApiController::sendWork,    "/api/send-work",   drogon::Get);
    METHOD_LIST_END

    void getAll(const drogon::HttpRequestPtr& req,
                std::function<void(const drogon::HttpResponsePtr&)>&& callback);
    void getById(const drogon::HttpRequestPtr& req,
                 std::function<void(const drogon::HttpResponsePtr&)>&& callback,
                 int id);
    void create(const drogon::HttpRequestPtr& req,
                std::function<void(const drogon::HttpResponsePtr&)>&& callback);
    void updateById(const drogon::HttpRequestPtr& req,
                    std::function<void(const drogon::HttpResponsePtr&)>&& callback,
                    int id);
    void deleteById(const drogon::HttpRequestPtr& req,
                    std::function<void(const drogon::HttpResponsePtr&)>&& callback,
                    int id);
    void sendWork(const drogon::HttpRequestPtr& req,
                  std::function<void(const drogon::HttpResponsePtr&)>&& callback);

    void setService(std::shared_ptr<ApiService> svc);
    void setEventBus(std::shared_ptr<EventBus> bus);

private:
    std::shared_ptr<ApiService> svc_;
    std::shared_ptr<EventBus>   bus_;
};

} // namespace app
