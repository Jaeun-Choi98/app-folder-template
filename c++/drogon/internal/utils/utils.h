#pragma once

#include <cstdint>
#include <string>
#include <vector>

namespace app::utils {

std::string bytesToHex(const std::vector<uint8_t>& data);
std::vector<uint8_t> hexToBytes(const std::string& hex);
std::string currentTimestamp();

} // namespace app::utils
