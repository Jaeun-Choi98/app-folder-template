package juchoi.template.appfolder.transport.tcp.server.parser;

import java.io.IOException;
import java.io.InputStream;

/**
 * Parses binary frames from a TCP InputStream.
 * Frame format: [1 byte opcode][4 bytes big-endian body length][N bytes body]
 * Mirrors Go's server/parser/rtms_parser.go.
 */
public class RtmsParser {

    public record Frame(byte opcode, byte[] body) {}

    public Frame parse(InputStream in) throws IOException {
        int opcode = in.read();
        if (opcode == -1) return null; // peer closed connection

        byte[] lenBuf = in.readNBytes(4);
        if (lenBuf.length < 4) return null;

        int length = ((lenBuf[0] & 0xFF) << 24)
                   | ((lenBuf[1] & 0xFF) << 16)
                   | ((lenBuf[2] & 0xFF) << 8)
                   |  (lenBuf[3] & 0xFF);

        byte[] body = in.readNBytes(length);
        return new Frame((byte) opcode, body);
    }
}
