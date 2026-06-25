#pragma once

#include <cstdint>
#include <string>
#include <vector>

namespace app {

class BinaryWriter {
public:
    // [1B opcode][4B big-endian length][NB body]
    static std::vector<uint8_t> build(uint8_t opcode,
                                      const std::string& body);
};

} // namespace app
