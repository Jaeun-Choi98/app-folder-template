#include "internal/transport/http_rest/controller/api_controller.h"
#include "internal/transport/http_rest/response/base_response.h"
#include "internal/service/service.h"
#include "internal/event/event_bus.h"
#include "internal/event/event_define.h"

namespace app {

void ApiController::setService(std::shared_ptr<ApiService> svc) {
    svc_ = std::move(svc);
}
void ApiController::setEventBus(std::shared_ptr<EventBus> bus) {
    bus_ = std::move(bus);
}

void ApiController::getAll(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback) {
    auto entities = svc_->getAll();
    Json::Value arr(Json::arrayValue);
    for (auto& e : entities) {
        Json::Value v;
        v["id"]    = e.id;
        v["name"]  = e.name;
        v["value"] = e.value;
        arr.append(v);
    }
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(arr)));
}

void ApiController::getById(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback,
    int id) {
    auto entity = svc_->getById(id);
    if (!entity) {
        auto resp = drogon::HttpResponse::newHttpJsonResponse(
            makeErrorResponse(404, "not found"));
        resp->setStatusCode(drogon::k404NotFound);
        callback(resp);
        return;
    }
    Json::Value v;
    v["id"]    = entity->id;
    v["name"]  = entity->name;
    v["value"] = entity->value;
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(v)));
}

void ApiController::create(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback) {
    auto json = req->getJsonObject();
    if (!json) {
        auto resp = drogon::HttpResponse::newHttpJsonResponse(
            makeErrorResponse(400, "invalid json"));
        resp->setStatusCode(drogon::k400BadRequest);
        callback(resp);
        return;
    }
    auto id = svc_->create((*json)["name"].asString(),
                           (*json)["value"].asString());
    Json::Value v;
    v["id"] = id;
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(v)));
}

void ApiController::updateById(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback,
    int id) {
    auto json = req->getJsonObject();
    if (!json) {
        auto resp = drogon::HttpResponse::newHttpJsonResponse(
            makeErrorResponse(400, "invalid json"));
        resp->setStatusCode(drogon::k400BadRequest);
        callback(resp);
        return;
    }
    bool ok = svc_->update(id, (*json)["name"].asString(),
                           (*json)["value"].asString());
    Json::Value v;
    v["updated"] = ok;
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(v)));
}

void ApiController::deleteById(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback,
    int id) {
    bool ok = svc_->remove(id);
    Json::Value v;
    v["deleted"] = ok;
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(v)));
}

void ApiController::sendWork(
    const drogon::HttpRequestPtr& req,
    std::function<void(const drogon::HttpResponsePtr&)>&& callback) {
    // TODO: EventBus를 통해 TCP 전송 이벤트 발행
    Json::Value v;
    v["sent"] = true;
    callback(drogon::HttpResponse::newHttpJsonResponse(
        makeSuccessResponse(v)));
}

} // namespace app
