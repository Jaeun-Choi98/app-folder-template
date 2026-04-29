package juchoi.template.appfolder.transport.tcp.server;

import juchoi.template.appfolder.eventbus.EventBus;
import juchoi.template.appfolder.eventbus.event.TcpSendEvent;
import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.transport.tcp.server.client.ClientManager;
import juchoi.template.appfolder.transport.tcp.server.client.TcpClient;
import juchoi.template.appfolder.transport.tcp.server.serializer.BinaryWriter;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.net.ServerSocket;
import java.net.Socket;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicInteger;

// TcpConfig에서 @Bean으로 등록
@Slf4j
@RequiredArgsConstructor
public class TcpServer {

    private final int port;
    private final ClientManager clientManager;
    private final TcpService tcpService;
    private final EventBus eventBus;

    private ServerSocket serverSocket;
    private final Executor executor = Executors.newCachedThreadPool();
    private final AtomicInteger idGen = new AtomicInteger(1);
    private volatile boolean running = false;

    public void start() {
        eventBus.subscribe("tcpclient.send", (TcpSendEvent event) -> {
            TcpClient client = clientManager.get(event.getClientId());
            if (client == null) {
                log.warn("[TCP] client {} not found", event.getClientId());
                event.fail(new IllegalStateException("client not found"));
                return;
            }
            try {
                byte[] frame = BinaryWriter.build(event.getOpcode(), event.getData());
                client.send(frame);
                event.complete(null);
            } catch (IOException e) {
                log.error("[TCP] send failed to client {}", event.getClientId(), e);
                event.fail(e);
            }
        });

        try {
            serverSocket = new ServerSocket(port);
            running = true;
            log.info("[TCP] Listening on port {}", port);

            CompletableFuture
                .runAsync(this::acceptLoop, executor)
                .exceptionally(e -> {
                    log.error("[TCP] acceptLoop terminated unexpectedly", e);
                    return null;
                });
        } catch (IOException e) {
            log.error("[TCP] Failed to bind port {}", port, e);
        }
    }

    private void acceptLoop() {
        while (running) {
            try {
                Socket socket = serverSocket.accept();
                int id = idGen.getAndIncrement();
                TcpClient client = new TcpClient(id, socket, tcpService, clientManager);
                clientManager.register(id, client);

                CompletableFuture
                    .runAsync(client::readLoop, executor)
                    .exceptionally(e -> {
                        log.error("[TCP] readLoop error for client {}", id, e);
                        clientManager.unregister(id);
                        return null;
                    });

                log.info("[TCP] Client connected id={}", id);
            } catch (IOException e) {
                if (running) log.error("[TCP] Accept error", e);
            }
        }
    }

    public void shutdown() {
        running = false;
        try {
            if (serverSocket != null) serverSocket.close();
        } catch (IOException ignored) {}
        ((ExecutorService) executor).shutdown();
        try {
            ((ExecutorService) executor).awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException ignored) {}
        log.info("[TCP] Server shut down");
    }
}
