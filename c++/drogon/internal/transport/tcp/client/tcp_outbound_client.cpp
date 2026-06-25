#include "internal/transport/tcp/client/tcp_outbound_client.h"
#include "internal/logger/logger.h"

namespace app {

TcpOutboundClient::TcpOutboundClient(asio::io_context& ioc)
    : socket_(ioc) {}

bool TcpOutboundClient::connect(const std::string& host, int port) {
    try {
        asio::ip::tcp::resolver resolver(socket_.get_executor());
        auto endpoints = resolver.resolve(host, std::to_string(port));
        asio::connect(socket_, endpoints);
        connected_ = true;
        return true;
    } catch (const std::exception& e) {
        Logger::instance().error(
            std::string("tcp outbound connect failed: ") + e.what());
        return false;
    }
}

void TcpOutboundClient::disconnect() {
    std::lock_guard lock(mtx_);
    if (connected_) {
        asio::error_code ec;
        socket_.close(ec);
        connected_ = false;
    }
}

bool TcpOutboundClient::isConnected() const { return connected_; }

void TcpOutboundClient::send(const std::vector<uint8_t>& data) {
    std::lock_guard lock(mtx_);
    if (!connected_) return;
    asio::error_code ec;
    asio::write(socket_, asio::buffer(data), ec);
    if (ec) {
        Logger::instance().error("tcp outbound send failed");
    }
}

} // namespace app
