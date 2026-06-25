#include "internal/event/event_bus.h"

namespace app {

void EventBus::subscribe(const std::string& topic, EventHandler handler) {
    std::lock_guard lock(mtx_);
    handlers_[topic].push_back(std::move(handler));
}

void EventBus::publish(const std::string& topic, const std::any& payload) {
    std::vector<EventHandler> snapshot;
    {
        std::lock_guard lock(mtx_);
        auto it = handlers_.find(topic);
        if (it == handlers_.end()) return;
        snapshot = it->second;
    }
    for (auto& h : snapshot) {
        h(payload);
    }
}

void EventBus::clear() {
    std::lock_guard lock(mtx_);
    handlers_.clear();
}

} // namespace app
