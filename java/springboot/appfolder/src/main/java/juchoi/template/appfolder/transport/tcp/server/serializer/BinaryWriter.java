package juchoi.template.appfolder.transport.tcp.server.serializer;

import juchoi.template.appfolder.utils.ByteUtils;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.nio.ByteOrder;

// Packet format: [opcode 1B][length 2B big-endian][data length-B]
public class BinaryWriter {

    public static byte[] build(byte opcode, byte[] data) throws IOException {
        ByteArrayOutputStream buf = new ByteArrayOutputStream(3 + data.length);
        buf.write(opcode);
        buf.write(ByteUtils.fromShort((short) data.length, ByteOrder.BIG_ENDIAN));
        buf.write(data);
        return buf.toByteArray();
    }
}
