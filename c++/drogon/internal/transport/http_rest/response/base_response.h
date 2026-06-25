#pragma once

#include <drogon/HttpResponse.h>
#include <json/json.h>
#include <string>

namespace app {

inline Json::Value makeSuccessResponse(const Json::Value& data) {
    Json::Value root;
    root["success"] = true;
    root["data"]    = data;
    return root;
}

inline Json::Value makeErrorResponse(int code, const std::string& message) {
    Json::Value root;
    root["success"] = false;
    root["error"]["code"]    = code;
    root["error"]["message"] = message;
    return root;
}

} // namespace app
