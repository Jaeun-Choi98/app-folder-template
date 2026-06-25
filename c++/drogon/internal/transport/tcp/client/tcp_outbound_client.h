#pragma once

#include <asio.hpp>
#include <cstdint>
#include <mutex>
#include <string>
#include <vector>

namespace app {

class TcpOutboundClient {
public:
    TcpOutboundClient(asio::io_context& ioc);

    bool connect(const std::string& host, int port);
    void disconnect();
    bool isConnected() const;

    void send(const std::vector<uint8_t>& data);

private:
    asio::ip::tcp::socket socket_;
    std::mutex            mtx_;
    bool                  connected_ = false;
};

} // namespace app
