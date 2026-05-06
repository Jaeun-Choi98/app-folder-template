package juchoi.template.appfolder.transport.tcp.server.dispatcher;

import juchoi.template.appfolder.service.TcpService;
import juchoi.template.appfolder.transport.tcp.server.parser.PacketParser.Packet;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@RequiredArgsConstructor
public class PacketDispatcher {

    private final int clientId;
    private final TcpService tcpService;

    private final ConcurrentHashMap<Byte, CompletableFuture<byte[]>> pendingReplies = new ConcurrentHashMap<>();

    public void dispatch(Packet packet) {
        CompletableFuture<byte[]> pending = pendingReplies.remove(packet.opcode());
        if (pending != null) {
            pending.complete(packet.data());
        }
        switch (packet.opcode()) {
            case 0x21 -> tcpService.handle0x21(clientId);
            case (byte) 0x22 -> tcpService.handle0x22(clientId);
            default -> log.warn("[TCP] Unknown opcode 0x{} from client {}",
                    Integer.toHexString(packet.opcode() & 0xFF), clientId);
        }
    }

    public void registerPending(byte replyOpcode, CompletableFuture<byte[]> future) {
        pendingReplies.put(replyOpcode, future);
    }

    public void failAll(Throwable cause) {
        pendingReplies.values().forEach(f -> f.completeExceptionally(cause));
        pendingReplies.clear();
    }
}
