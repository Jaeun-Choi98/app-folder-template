#pragma once

#include "internal/db/entity/sample_entity.h"

#include <optional>
#include <string>
#include <vector>

namespace app {

class ApiService {
public:
    virtual ~ApiService() = default;

    virtual std::vector<SampleEntity> getAll() = 0;
    virtual std::optional<SampleEntity> getById(int id) = 0;
    virtual int create(const std::string& name, const std::string& value) = 0;
    virtual bool update(int id, const std::string& name, const std::string& value) = 0;
    virtual bool remove(int id) = 0;
};

class TcpService {
public:
    virtual ~TcpService() = default;

    virtual void handleFrame(uint32_t clientId, uint8_t opcode,
                             const std::string& body) = 0;
};

} // namespace app
