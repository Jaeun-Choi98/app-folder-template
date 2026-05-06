package juchoi.template.appfolder.eventbus.event;

import java.util.concurrent.CompletableFuture;

import lombok.Getter;

/**
 * Published by REST handlers on "tcpclient.send", consumed by TcpWorker.
 * Mirrors Go's TCPSendPayload struct + Response chan.
 *
 * Go uses a channel: Response chan TCPResponse
 * Java uses: CompletableFuture<byte[]>
 * → same semantics: publisher blocks on future.get(), worker completes it.
 *
 * replyOpcode: 응답을 기다리는 경우 예상 수신 opcode 지정 (null이면 fire-and-forget)
 */
@Getter
public class TcpSendEvent {

    private final int clientId;
    private final byte opcode;
    private final byte[] data;
    private final Byte replyOpcode;

    private final CompletableFuture<byte[]> responseFuture = new CompletableFuture<>();

    private TcpSendEvent(int clientId, byte opcode, byte[] data, Byte replyOpcode) {
        this.clientId = clientId;
        this.opcode = opcode;
        this.data = data;
        this.replyOpcode = replyOpcode;
    }

    // Fire-and-forget: 응답 대기 없음
    public static TcpSendEvent noReply(int clientId, byte opcode, byte[] data) {
        return new TcpSendEvent(clientId, opcode, data, null);
    }

    // 응답 대기: replyOpcode 수신 시 future 완료
    public static TcpSendEvent withReply(int clientId, byte opcode, byte[] data, byte replyOpcode) {
        return new TcpSendEvent(clientId, opcode, data, replyOpcode);
    }

    public boolean expectsReply() {
        return replyOpcode != null;
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
