#pragma once

#include <asio.hpp>
#include <atomic>
#include <cstdint>
#include <memory>
#include <thread>

namespace app {

struct Config;
class ClientManager;
class TcpService;

class TcpServer {
public:
    TcpServer(const Config& cfg,
              std::shared_ptr<ClientManager> mgr,
              std::shared_ptr<TcpService> svc);

    void start();
    void stop();

    asio::io_context& ioContext() { return ioc_; }

private:
    void doAccept();

    asio::io_context        ioc_;
    asio::ip::tcp::acceptor acceptor_;
    std::thread             thread_;

    std::shared_ptr<ClientManager> mgr_;
    std::shared_ptr<TcpService>    svc_;
    std::atomic<uint32_t>          nextId_{1};
};

} // namespace app
