package juchoi.template.appfolder.eventbus.event;

import java.util.concurrent.CompletableFuture;

import lombok.Getter;
import lombok.RequiredArgsConstructor;

/**
 * Published by REST handlers on "tcpclient.send", consumed by TcpWorker.
 * Mirrors Go's TCPSendPayload struct + Response chan.
 *
 * Go uses a channel: Response chan TCPResponse
 * Java uses: CompletableFuture<byte[]>
 * → same semantics: publisher blocks on future.get(), worker completes it.
 */
@Getter
@RequiredArgsConstructor
public class TcpSendEvent {

    private final int clientId;
    private final byte opcode;
    private final byte[] data;

    private final CompletableFuture<byte[]> responseFuture = new CompletableFuture<>();

    // Fire-and-forget shorthand (no need to await reply)
    public static TcpSendEvent noReply(int clientId, byte opcode, byte[] data) {
        return new TcpSendEvent(clientId, opcode, data);
    }

    public void complete(byte[] reply) {
        responseFuture.complete(reply);
    }

    public void fail(Throwable cause) {
        responseFuture.completeExceptionally(cause);
    }

    public CompletableFuture<byte[]> future() {
        return responseFuture;
    }
}
