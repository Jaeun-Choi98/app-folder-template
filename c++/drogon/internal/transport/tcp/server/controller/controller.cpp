#include "internal/transport/tcp/server/controller/controller.h"
#include "internal/service/service.h"
#include "internal/transport/tcp/server/client/client_manager.h"

namespace app {

TcpController::TcpController(std::shared_ptr<TcpService> svc,
                             std::shared_ptr<ClientManager> mgr)
    : svc_(std::move(svc)), mgr_(std::move(mgr)) {}

void TcpController::dispatch(uint32_t clientId, uint8_t opcode,
                             const std::string& body) {
    svc_->handleFrame(clientId, opcode, body);
}

} // namespace app
