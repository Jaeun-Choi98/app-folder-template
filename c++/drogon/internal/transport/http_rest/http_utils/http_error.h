#pragma once

#include <drogon/HttpResponse.h>
#include <json/json.h>
#include <string>

namespace app {

inline drogon::HttpResponsePtr makeHttpError(
    drogon::HttpStatusCode code, const std::string& message) {
    Json::Value body;
    body["success"]          = false;
    body["error"]["code"]    = static_cast<int>(code);
    body["error"]["message"] = message;
    auto resp = drogon::HttpResponse::newHttpJsonResponse(body);
    resp->setStatusCode(code);
    return resp;
}

} // namespace app
