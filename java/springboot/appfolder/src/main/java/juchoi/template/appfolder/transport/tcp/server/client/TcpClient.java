package juchoi.template.appfolder.transport.tcp.server.client;

import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.transport.tcp.server.dispatcher.PacketDispatcher;
import juchoi.template.appfolder.transport.tcp.server.parser.PacketParser;
import juchoi.template.appfolder.transport.tcp.server.parser.PacketParser.Packet;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.util.concurrent.CompletableFuture;

@Slf4j
public class TcpClient {

    private final int id;
    private final Socket socket;
    private final OutputStream out;
    private final InputStream in;
    private final ClientManager clientManager;
    private final PacketParser parser = new PacketParser();
    private final PacketDispatcher dispatcher;

    public TcpClient(int id, Socket socket, TcpService tcpService, ClientManager clientManager)
            throws IOException {
        this.id = id;
        this.socket = socket;
        this.out = socket.getOutputStream();
        this.in = socket.getInputStream();
        this.clientManager = clientManager;
        this.dispatcher = new PacketDispatcher(id, tcpService);
    }

    // Blocking read loop — runs on its own thread (submitted by TcpServer).
    public void readLoop() {
        try {
            while (!socket.isClosed()) {
                Packet packet = parser.parse(in);
                if (packet == null) break;
                dispatcher.dispatch(packet);
            }
        } catch (IOException e) {
            log.info("[TCP] Client {} disconnected: {}", id, e.getMessage());
        } finally {
            dispatcher.failAll(new IOException("client disconnected"));
            clientManager.unregister(id);
            close();
        }
    }

    public void registerPending(byte replyOpcode, CompletableFuture<byte[]> future) {
        dispatcher.registerPending(replyOpcode, future);
    }

    // Thread-safe write — multiple workers may call this concurrently.
    public synchronized void send(byte[] data) throws IOException {
        out.write(data);
        out.flush();
    }

    public int getId() { return id; }

    public void close() {
        try { socket.close(); } catch (IOException ignored) {}
    }
}
