#pragma once

#include <json/json.h>
#include <string>

namespace app {

struct SampleResponse {
    int         id;
    std::string name;
    std::string value;

    Json::Value toJson() const {
        Json::Value v;
        v["id"]    = id;
        v["name"]  = name;
        v["value"] = value;
        return v;
    }
};

} // namespace app
