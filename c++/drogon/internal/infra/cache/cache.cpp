#include "internal/infra/cache/cache.h"

namespace app {

void Cache::set(const std::string& key, const std::any& value) {
    std::lock_guard lock(mtx_);
    store_[key] = value;
    dirty_ = true;
}

std::optional<std::any> Cache::get(const std::string& key) const {
    std::lock_guard lock(mtx_);
    auto it = store_.find(key);
    if (it == store_.end()) return std::nullopt;
    return it->second;
}

bool Cache::remove(const std::string& key) {
    std::lock_guard lock(mtx_);
    auto erased = store_.erase(key);
    if (erased > 0) dirty_ = true;
    return erased > 0;
}

void Cache::clear() {
    std::lock_guard lock(mtx_);
    store_.clear();
    dirty_ = true;
}

std::vector<std::string> Cache::keys() const {
    std::lock_guard lock(mtx_);
    std::vector<std::string> result;
    result.reserve(store_.size());
    for (auto& [k, _] : store_)
        result.push_back(k);
    return result;
}

size_t Cache::size() const {
    std::lock_guard lock(mtx_);
    return store_.size();
}

void Cache::loadFromEntries(
    const std::vector<std::pair<std::string, std::any>>& entries) {
    std::lock_guard lock(mtx_);
    store_.clear();
    for (auto& [k, v] : entries)
        store_[k] = v;
    dirty_ = false;
}

bool Cache::isDirty() const {
    std::lock_guard lock(mtx_);
    return dirty_;
}

void Cache::resetDirty() {
    std::lock_guard lock(mtx_);
    dirty_ = false;
}

} // namespace app
