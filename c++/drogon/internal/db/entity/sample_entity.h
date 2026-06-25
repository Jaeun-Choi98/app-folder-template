#pragma once

#include <string>

namespace app {

struct SampleEntity {
    int         id   = 0;
    std::string name;
    std::string value;
};

struct LogEntity {
    int         id = 0;
    std::string level;
    std::string message;
    std::string createdAt;
};

} // namespace app
