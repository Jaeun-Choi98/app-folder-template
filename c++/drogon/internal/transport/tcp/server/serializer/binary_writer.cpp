#include "internal/transport/tcp/server/serializer/binary_writer.h"

namespace app {

std::vector<uint8_t> BinaryWriter::build(uint8_t opcode,
                                         const std::string& body) {
    uint32_t len = static_cast<uint32_t>(body.size());
    std::vector<uint8_t> buf;
    buf.reserve(5 + len);

    buf.push_back(opcode);
    buf.push_back(static_cast<uint8_t>((len >> 24) & 0xFF));
    buf.push_back(static_cast<uint8_t>((len >> 16) & 0xFF));
    buf.push_back(static_cast<uint8_t>((len >> 8)  & 0xFF));
    buf.push_back(static_cast<uint8_t>((len)       & 0xFF));
    buf.insert(buf.end(), body.begin(), body.end());

    return buf;
}

} // namespace app
