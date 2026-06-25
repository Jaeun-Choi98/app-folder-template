#include "internal/transport/tcp/server/tcp_client.h"
#include "internal/transport/tcp/server/parser/rtms_parser.h"
#include "internal/logger/logger.h"

#include <array>

namespace app {

TcpClient::TcpClient(asio::ip::tcp::socket socket, uint32_t id)
    : socket_(std::move(socket)), id_(id) {}

void TcpClient::start(FrameHandler onFrame) {
    onFrame_ = std::move(onFrame);
    doRead();
}

void TcpClient::doRead() {
    auto self = shared_from_this();
    auto buf  = std::make_shared<std::array<uint8_t, 4096>>();

    socket_.async_read_some(
        asio::buffer(*buf),
        [this, self, buf](asio::error_code ec, std::size_t n) {
            if (ec) {
                Logger::instance().info(
                    "tcp client " + std::to_string(id_) + " disconnected");
                return;
            }

            readBuf_.insert(readBuf_.end(), buf->begin(), buf->begin() + n);

            RtmsParser parser;
            size_t consumed = 0;
            while (auto frame = parser.parse(readBuf_, consumed)) {
                if (onFrame_) onFrame_(id_, *frame);
                readBuf_.erase(readBuf_.begin(),
                               readBuf_.begin() + consumed);
                consumed = 0;
            }

            doRead();
        });
}

void TcpClient::send(const std::vector<uint8_t>& data) {
    std::lock_guard lock(writeMtx_);
    asio::error_code ec;
    asio::write(socket_, asio::buffer(data), ec);
    if (ec) {
        Logger::instance().error(
            "tcp send failed for client " + std::to_string(id_));
    }
}

void TcpClient::close() {
    asio::error_code ec;
    socket_.close(ec);
}

} // namespace app
