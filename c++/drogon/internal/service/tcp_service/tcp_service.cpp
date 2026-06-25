#include "internal/service/tcp_service/tcp_service.h"
#include "internal/db/db_handler/db_handler.h"
#include "internal/infra/cache/cache.h"

namespace app {

TcpServiceImpl::TcpServiceImpl(std::shared_ptr<DbHandler> db,
                               std::shared_ptr<Cache> cache)
    : db_(std::move(db))
    , cache_(std::move(cache)) {}

void TcpServiceImpl::handleFrame(uint32_t clientId, uint8_t opcode,
                                 const std::string& body) {
    // TODO: opcode에 따라 분기 처리
    // 예: opcode 0x01 → 캐시 업데이트
    //     opcode 0x02 → DB 저장
}

} // namespace app
