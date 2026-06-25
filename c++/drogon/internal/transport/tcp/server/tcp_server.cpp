#include "internal/transport/tcp/server/tcp_server.h"
#include "internal/transport/tcp/server/tcp_client.h"
#include "internal/transport/tcp/server/client/client_manager.h"
#include "internal/config/config.h"
#include "internal/service/service.h"
#include "internal/logger/logger.h"

namespace app {

TcpServer::TcpServer(const Config& cfg,
                     std::shared_ptr<ClientManager> mgr,
                     std::shared_ptr<TcpService> svc)
    : acceptor_(ioc_,
                asio::ip::tcp::endpoint(
                    asio::ip::make_address(cfg.tcpHost.empty() ? "0.0.0.0" : cfg.tcpHost),
                    cfg.tcpPort))
    , mgr_(std::move(mgr))
    , svc_(std::move(svc)) {}

void TcpServer::start() {
    doAccept();
    thread_ = std::thread([this] { ioc_.run(); });
    Logger::instance().info(
        "tcp server started on port " +
        std::to_string(acceptor_.local_endpoint().port()));
}

void TcpServer::stop() {
    ioc_.stop();
    if (thread_.joinable()) thread_.join();
}

void TcpServer::doAccept() {
    acceptor_.async_accept([this](asio::error_code ec,
                                  asio::ip::tcp::socket socket) {
        if (!ec) {
            auto id     = nextId_.fetch_add(1);
            auto client = std::make_shared<TcpClient>(std::move(socket), id);
            mgr_->add(id, client);

            auto svc = svc_;
            auto mgr = mgr_;
            client->start([svc, mgr](uint32_t cid, const TcpFrame& frame) {
                svc->handleFrame(cid, frame.opcode, frame.body);
            });

            Logger::instance().info(
                "tcp client connected: " + std::to_string(id));
        }
        doAccept();
    });
}

} // namespace app
