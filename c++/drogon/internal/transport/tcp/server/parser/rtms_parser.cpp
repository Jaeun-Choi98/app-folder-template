#include "internal/transport/tcp/server/parser/rtms_parser.h"
#include "internal/infra/define/define.h"

namespace app {

std::optional<TcpFrame> RtmsParser::parse(const std::vector<uint8_t>& buffer,
                                          size_t& consumed) {
    constexpr size_t kHeaderSize = 5; // 1B opcode + 4B length
    if (buffer.size() < kHeaderSize) return std::nullopt;

    uint8_t opcode = buffer[0];
    uint32_t length = (static_cast<uint32_t>(buffer[1]) << 24) |
                      (static_cast<uint32_t>(buffer[2]) << 16) |
                      (static_cast<uint32_t>(buffer[3]) << 8)  |
                      (static_cast<uint32_t>(buffer[4]));

    if (length > kTcpMaxFrameSize) return std::nullopt;
    if (buffer.size() < kHeaderSize + length) return std::nullopt;

    TcpFrame frame;
    frame.opcode = opcode;
    frame.length = length;
    frame.body.assign(buffer.begin() + kHeaderSize,
                      buffer.begin() + kHeaderSize + length);

    consumed = kHeaderSize + length;
    return frame;
}

} // namespace app
