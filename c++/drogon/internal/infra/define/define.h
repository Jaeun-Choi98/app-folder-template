#pragma once

#include <cstdint>
#include <string>

namespace app {

constexpr int    kCacheWorkerIntervalMs = 1000;
constexpr int    kLogCleanupDays        = 3;
constexpr size_t kTcpMaxFrameSize       = 1024 * 1024; // 1MB

} // namespace app
