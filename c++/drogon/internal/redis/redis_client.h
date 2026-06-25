#pragma once

#include <string>
#include <optional>
#include <mutex>

struct redisContext;

namespace app {

struct Config;

class RedisClient {
public:
    RedisClient() = default;
    ~RedisClient();

    void connect(const Config& cfg);
    void disconnect();
    bool isConnected() const;

    bool setString(const std::string& key, const std::string& value);
    std::optional<std::string> getString(const std::string& key);
    bool del(const std::string& key);
    bool expire(const std::string& key, int seconds);

private:
    redisContext* ctx_ = nullptr;
    std::mutex    mtx_;
};

} // namespace app
