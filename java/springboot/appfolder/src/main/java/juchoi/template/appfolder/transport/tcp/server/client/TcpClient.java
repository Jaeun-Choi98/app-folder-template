package juchoi.template.appfolder.transport.tcp.server.client;

import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.transport.tcp.server.parser.RtmsParser;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;

// Represents a single connected TCP client. Mirrors Go's server/client/client.go.
@Slf4j
public class TcpClient {

    private final int id;
    private final Socket socket;
    private final OutputStream out;
    private final InputStream in;
    private final TcpService tcpService;
    private final ClientManager clientManager;
    private final RtmsParser parser = new RtmsParser();

    public TcpClient(int id, Socket socket, TcpService tcpService, ClientManager clientManager)
            throws IOException {
        this.id = id;
        this.socket = socket;
        this.out = socket.getOutputStream();
        this.in = socket.getInputStream();
        this.tcpService = tcpService;
        this.clientManager = clientManager;
    }

    // Blocking read loop — runs on its own thread (submitted by TcpServer).
    public void readLoop() {
        try {
            while (!socket.isClosed()) {
                RtmsParser.Frame frame = parser.parse(in);
                if (frame == null) break;
                dispatch(frame);
            }
        } catch (IOException e) {
            log.info("[TCP] Client {} disconnected: {}", id, e.getMessage());
        } finally {
            clientManager.unregister(id);
            close();
        }
    }

    private void dispatch(RtmsParser.Frame frame) {
        switch (frame.opcode()) {
            case 0x01       -> tcpService.handle0x01(id);
            case (byte)0xAA -> tcpService.handle0xAA(id);
            default -> log.warn("[TCP] Unknown opcode 0x{} from client {}", Integer.toHexString(frame.opcode()), id);
        }
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
