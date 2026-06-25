#pragma once

#include <cstdint>
#include <memory>
#include <mutex>
#include <unordered_map>

namespace app {

class TcpClient;

class ClientManager {
public:
    void add(uint32_t id, std::shared_ptr<TcpClient> client);
    void remove(uint32_t id);
    std::shared_ptr<TcpClient> get(uint32_t id) const;
    size_t count() const;
    std::vector<uint32_t> allIds() const;

private:
    mutable std::mutex mtx_;
    std::unordered_map<uint32_t, std::shared_ptr<TcpClient>> clients_;
};

} // namespace app
