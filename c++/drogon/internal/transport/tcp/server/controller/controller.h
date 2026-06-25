#pragma once

#include <cstdint>
#include <memory>
#include <string>

namespace app {

class TcpService;
class ClientManager;

class TcpController {
public:
    TcpController(std::shared_ptr<TcpService> svc,
                  std::shared_ptr<ClientManager> mgr);

    void dispatch(uint32_t clientId, uint8_t opcode,
                  const std::string& body);

private:
    std::shared_ptr<TcpService>    svc_;
    std::shared_ptr<ClientManager> mgr_;
};

} // namespace app
