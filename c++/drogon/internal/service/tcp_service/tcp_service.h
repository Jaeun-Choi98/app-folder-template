#pragma once

#include "internal/service/service.h"

#include <memory>

namespace app {

class DbHandler;
class Cache;

class TcpServiceImpl : public TcpService {
public:
    TcpServiceImpl(std::shared_ptr<DbHandler> db,
                   std::shared_ptr<Cache> cache);

    void handleFrame(uint32_t clientId, uint8_t opcode,
                     const std::string& body) override;

private:
    std::shared_ptr<DbHandler> db_;
    std::shared_ptr<Cache>     cache_;
};

} // namespace app
