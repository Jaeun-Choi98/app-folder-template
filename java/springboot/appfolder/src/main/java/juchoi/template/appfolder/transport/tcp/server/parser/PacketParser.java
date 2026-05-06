package juchoi.template.appfolder.transport.tcp.server.parser;

import juchoi.template.appfolder.utils.ByteUtils;

import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteOrder;

// Packet format: [opcode 1B][length 2B big-endian][data length-B]
public class PacketParser {

    public record Packet(byte opcode, byte[] data) {
    }

    public Packet parse(InputStream in) throws IOException {
        int opcode = in.read();
        if (opcode == -1)
            return null;

        byte[] lenBuf = in.readNBytes(2);
        if (lenBuf.length < 2)
            return null;

        int length = ByteUtils.toShort(lenBuf, ByteOrder.BIG_ENDIAN);
        byte[] data = in.readNBytes(length);
        return new Packet((byte) opcode, data);
    }
}
