#include "internal/service/api_service/api_service.h"
#include "internal/db/db_handler/db_handler.h"
#include "internal/infra/cache/cache.h"
#include "internal/event/event_bus.h"
#include "internal/redis/redis_client.h"

namespace app {

ApiServiceImpl::ApiServiceImpl(std::shared_ptr<DbHandler> db,
                               std::shared_ptr<Cache> cache,
                               std::shared_ptr<EventBus> bus,
                               std::shared_ptr<RedisClient> redis)
    : db_(std::move(db))
    , cache_(std::move(cache))
    , bus_(std::move(bus))
    , redis_(std::move(redis)) {}

std::vector<SampleEntity> ApiServiceImpl::getAll() {
    return db_->findAll();
}

std::optional<SampleEntity> ApiServiceImpl::getById(int id) {
    return db_->findById(id);
}

int ApiServiceImpl::create(const std::string& name, const std::string& value) {
    SampleEntity e;
    e.name  = name;
    e.value = value;
    return db_->insert(e);
}

bool ApiServiceImpl::update(int id, const std::string& name,
                            const std::string& value) {
    SampleEntity e;
    e.id    = id;
    e.name  = name;
    e.value = value;
    return db_->update(e);
}

bool ApiServiceImpl::remove(int id) {
    return db_->remove(id);
}

} // namespace app
