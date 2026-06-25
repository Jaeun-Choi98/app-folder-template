#include "internal/redis/redis_client.h"
#include "internal/config/config.h"

#include <hiredis/hiredis.h>
#include <stdexcept>

namespace app {

RedisClient::~RedisClient() { disconnect(); }

void RedisClient::connect(const Config& cfg) {
    std::lock_guard lock(mtx_);
    ctx_ = redisConnect(cfg.redisHost.c_str(), cfg.redisPort);
    if (!ctx_ || ctx_->err)
        throw std::runtime_error("redis connect failed");

    if (!cfg.redisPassword.empty()) {
        auto* reply = static_cast<redisReply*>(
            redisCommand(ctx_, "AUTH %s", cfg.redisPassword.c_str()));
        if (reply) freeReplyObject(reply);
    }
}

void RedisClient::disconnect() {
    std::lock_guard lock(mtx_);
    if (ctx_) { redisFree(ctx_); ctx_ = nullptr; }
}

bool RedisClient::isConnected() const { return ctx_ != nullptr; }

bool RedisClient::setString(const std::string& key, const std::string& value) {
    std::lock_guard lock(mtx_);
    if (!ctx_) return false;
    auto* reply = static_cast<redisReply*>(
        redisCommand(ctx_, "SET %s %s", key.c_str(), value.c_str()));
    if (!reply) return false;
    bool ok = (reply->type == REDIS_REPLY_STATUS);
    freeReplyObject(reply);
    return ok;
}

std::optional<std::string> RedisClient::getString(const std::string& key) {
    std::lock_guard lock(mtx_);
    if (!ctx_) return std::nullopt;
    auto* reply = static_cast<redisReply*>(
        redisCommand(ctx_, "GET %s", key.c_str()));
    if (!reply) return std::nullopt;
    std::optional<std::string> result;
    if (reply->type == REDIS_REPLY_STRING)
        result = std::string(reply->str, reply->len);
    freeReplyObject(reply);
    return result;
}

bool RedisClient::del(const std::string& key) {
    std::lock_guard lock(mtx_);
    if (!ctx_) return false;
    auto* reply = static_cast<redisReply*>(
        redisCommand(ctx_, "DEL %s", key.c_str()));
    if (!reply) return false;
    bool ok = (reply->type == REDIS_REPLY_INTEGER && reply->integer > 0);
    freeReplyObject(reply);
    return ok;
}

bool RedisClient::expire(const std::string& key, int seconds) {
    std::lock_guard lock(mtx_);
    if (!ctx_) return false;
    auto* reply = static_cast<redisReply*>(
        redisCommand(ctx_, "EXPIRE %s %d", key.c_str(), seconds));
    if (!reply) return false;
    bool ok = (reply->type == REDIS_REPLY_INTEGER && reply->integer == 1);
    freeReplyObject(reply);
    return ok;
}

} // namespace app
