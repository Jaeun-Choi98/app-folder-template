#pragma once

#include "internal/service/service.h"

#include <memory>

namespace app {

class DbHandler;
class Cache;
class EventBus;
class RedisClient;

class ApiServiceImpl : public ApiService {
public:
    ApiServiceImpl(std::shared_ptr<DbHandler> db,
                   std::shared_ptr<Cache> cache,
                   std::shared_ptr<EventBus> bus,
                   std::shared_ptr<RedisClient> redis);

    std::vector<SampleEntity>   getAll() override;
    std::optional<SampleEntity> getById(int id) override;
    int  create(const std::string& name, const std::string& value) override;
    bool update(int id, const std::string& name, const std::string& value) override;
    bool remove(int id) override;

private:
    std::shared_ptr<DbHandler>   db_;
    std::shared_ptr<Cache>       cache_;
    std::shared_ptr<EventBus>    bus_;
    std::shared_ptr<RedisClient> redis_;
};

} // namespace app
