#include "internal/transport/tcp/server/client/client_manager.h"

namespace app {

void ClientManager::add(uint32_t id, std::shared_ptr<TcpClient> client) {
    std::lock_guard lock(mtx_);
    clients_[id] = std::move(client);
}

void ClientManager::remove(uint32_t id) {
    std::lock_guard lock(mtx_);
    clients_.erase(id);
}

std::shared_ptr<TcpClient> ClientManager::get(uint32_t id) const {
    std::lock_guard lock(mtx_);
    auto it = clients_.find(id);
    return it != clients_.end() ? it->second : nullptr;
}

size_t ClientManager::count() const {
    std::lock_guard lock(mtx_);
    return clients_.size();
}

std::vector<uint32_t> ClientManager::allIds() const {
    std::lock_guard lock(mtx_);
    std::vector<uint32_t> ids;
    ids.reserve(clients_.size());
    for (auto& [id, _] : clients_)
        ids.push_back(id);
    return ids;
}

} // namespace app
