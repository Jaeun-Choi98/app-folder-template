package juchoi.template.appfolder.transport.tcp.client;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.io.IOException;
import java.io.OutputStream;
import java.net.Socket;

/**
 * Outbound TCP client — this server connecting to an external TCP endpoint.
 * Mirrors Go's internal/transport/tcp/client/client.go.
 */
@Slf4j
@RequiredArgsConstructor
public class TcpOutboundClient {

    private final String host;
    private final int port;
    private Socket socket;
    private OutputStream out;

    public void connect() throws IOException {
        socket = new Socket(host, port);
        out = socket.getOutputStream();
        log.info("[TCP Client] Connected to {}:{}", host, port);
    }

    public synchronized void send(byte[] data) throws IOException {
        out.write(data);
        out.flush();
    }

    public void close() {
        try { if (socket != null) socket.close(); } catch (IOException ignored) {}
        log.info("[TCP Client] Disconnected from {}:{}", host, port);
    }
}
