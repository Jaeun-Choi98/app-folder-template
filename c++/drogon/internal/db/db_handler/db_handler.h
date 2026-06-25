#pragma once

#include "internal/db/entity/sample_entity.h"

#include <memory>
#include <optional>
#include <string>
#include <vector>

namespace app {

struct Config;

class DbHandler {
public:
    virtual ~DbHandler() = default;

    virtual void connect() = 0;
    virtual void disconnect() = 0;
    virtual bool isConnected() const = 0;

    // sample CRUD
    virtual std::vector<SampleEntity> findAll() = 0;
    virtual std::optional<SampleEntity> findById(int id) = 0;
    virtual int insert(const SampleEntity& entity) = 0;
    virtual bool update(const SampleEntity& entity) = 0;
    virtual bool remove(int id) = 0;

    // log
    virtual void insertLog(const LogEntity& log) = 0;

    static std::unique_ptr<DbHandler> create(const Config& cfg);
};

} // namespace app
