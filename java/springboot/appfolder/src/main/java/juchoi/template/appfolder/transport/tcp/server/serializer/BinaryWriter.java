package juchoi.template.appfolder.transport.tcp.server.serializer;

import java.io.ByteArrayOutputStream;
import java.io.IOException;

/**
 * Serializes outbound binary frames.
 * Frame format: [1 byte opcode][4 bytes big-endian body length][N bytes body]
 * Mirrors Go's server/serializer/binary_writer.go.
 */
public class BinaryWriter {

    public static byte[] build(byte opcode, byte[] body) throws IOException {
        ByteArrayOutputStream buf = new ByteArrayOutputStream(5 + body.length);
        buf.write(opcode);
        int len = body.length;
        buf.write((len >> 24) & 0xFF);
        buf.write((len >> 16) & 0xFF);
        buf.write((len >>  8) & 0xFF);
        buf.write(len & 0xFF);
        buf.write(body);
        return buf.toByteArray();
    }
}
