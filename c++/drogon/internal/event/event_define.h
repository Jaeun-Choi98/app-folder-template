#pragma once

#include <any>
#include <string>

namespace app {

// EventBus 토픽
namespace topic {
constexpr const char* kTcpNoReply = "tcpclient.send.noreply";
constexpr const char* kTcpWithReply = "tcpclient.send.reply";
constexpr const char* kCacheData = "cache.data";
constexpr const char* kSystemState = "system.state";
constexpr const char* kSystemEvent = "system.event";
}  // namespace topic

struct Event {
  std::string topic;
  std::any payload;
};

struct TcpSendEvent {
  uint32_t clientId = 0;
  uint8_t opcode = 0;
  std::string data;
};

struct CacheDataEvent {
  std::string key;
  std::string value;
};

}  // namespace app
