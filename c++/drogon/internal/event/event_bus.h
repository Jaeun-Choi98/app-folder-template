#pragma once

#include <any>
#include <functional>
#include <mutex>
#include <string>
#include <unordered_map>
#include <vector>

namespace app {

using EventHandler = std::function<void(const std::any&)>;

class EventBus {
public:
    void subscribe(const std::string& topic, EventHandler handler);
    void publish(const std::string& topic, const std::any& payload);
    void clear();

private:
    std::mutex mtx_;
    std::unordered_map<std::string, std::vector<EventHandler>> handlers_;
};

} // namespace app
