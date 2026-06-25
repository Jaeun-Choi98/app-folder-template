#pragma once

#include <any>
#include <mutex>
#include <optional>
#include <string>
#include <unordered_map>
#include <vector>

namespace app {

class Cache {
public:
    void set(const std::string& key, const std::any& value);
    std::optional<std::any> get(const std::string& key) const;
    bool remove(const std::string& key);
    void clear();
    std::vector<std::string> keys() const;
    size_t size() const;

    // DB에서 캐시를 로드할 때 사용
    void loadFromEntries(
        const std::vector<std::pair<std::string, std::any>>& entries);

    // 변경 추적
    bool isDirty() const;
    void resetDirty();

private:
    mutable std::mutex mtx_;
    std::unordered_map<std::string, std::any> store_;
    bool dirty_ = false;
};

} // namespace app
