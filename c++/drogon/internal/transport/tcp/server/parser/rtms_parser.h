#pragma once

#include <cstdint>
#include <optional>
#include <string>
#include <vector>

namespace app {

struct TcpFrame {
    uint8_t     opcode = 0;
    uint32_t    length = 0;
    std::string body;
};

class RtmsParser {
public:
    // 수신 버퍼에서 프레임을 파싱. 완전한 프레임이 없으면 nullopt 반환.
    // 형식: [1B opcode][4B big-endian length][NB body]
    std::optional<TcpFrame> parse(const std::vector<uint8_t>& buffer,
                                  size_t& consumed);
};

} // namespace app
