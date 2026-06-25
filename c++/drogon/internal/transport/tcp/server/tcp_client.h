#pragma once

#include <asio.hpp>
#include <cstdint>
#include <functional>
#include <memory>
#include <mutex>
#include <vector>

namespace app {

struct TcpFrame;

class TcpClient : public std::enable_shared_from_this<TcpClient> {
public:
    using FrameHandler = std::function<void(uint32_t, const TcpFrame&)>;

    TcpClient(asio::ip::tcp::socket socket, uint32_t id);

    void start(FrameHandler onFrame);
    void send(const std::vector<uint8_t>& data);
    void close();

    uint32_t id() const { return id_; }

private:
    void doRead();

    asio::ip::tcp::socket socket_;
    uint32_t              id_;
    FrameHandler          onFrame_;

    std::vector<uint8_t>  readBuf_;
    std::mutex            writeMtx_;
};

} // namespace app
