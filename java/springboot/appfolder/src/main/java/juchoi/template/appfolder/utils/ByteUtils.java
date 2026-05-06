package juchoi.template.appfolder.utils;

import java.nio.ByteOrder;

public class ByteUtils {

    // ── byte[] → numeric ──────────────────────────────────────────────────────

    public static short toShort(byte[] b, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return (short) (((b[0] & 0xFF) << 8) | (b[1] & 0xFF));
        else
            return (short) ((b[0] & 0xFF) | ((b[1] & 0xFF) << 8));
    }

    public static int toInt(byte[] b, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return ((b[0] & 0xFF) << 24) | ((b[1] & 0xFF) << 16)
                 | ((b[2] & 0xFF) <<  8) |  (b[3] & 0xFF);
        else
            return  (b[0] & 0xFF)        | ((b[1] & 0xFF) <<  8)
                 | ((b[2] & 0xFF) << 16) | ((b[3] & 0xFF) << 24);
    }

    public static long toLong(byte[] b, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return ((long)(b[0] & 0xFF) << 56) | ((long)(b[1] & 0xFF) << 48)
                 | ((long)(b[2] & 0xFF) << 40) | ((long)(b[3] & 0xFF) << 32)
                 | ((long)(b[4] & 0xFF) << 24) | ((long)(b[5] & 0xFF) << 16)
                 | ((long)(b[6] & 0xFF) <<  8) |  (long)(b[7] & 0xFF);
        else
            return  (long)(b[0] & 0xFF)        | ((long)(b[1] & 0xFF) <<  8)
                 | ((long)(b[2] & 0xFF) << 16) | ((long)(b[3] & 0xFF) << 24)
                 | ((long)(b[4] & 0xFF) << 32) | ((long)(b[5] & 0xFF) << 40)
                 | ((long)(b[6] & 0xFF) << 48) | ((long)(b[7] & 0xFF) << 56);
    }

    // ── numeric → byte[] ──────────────────────────────────────────────────────

    public static byte[] fromShort(short value, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return new byte[]{ (byte)(value >> 8), (byte) value };
        else
            return new byte[]{ (byte) value, (byte)(value >> 8) };
    }

    public static byte[] fromInt(int value, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return new byte[]{
                (byte)(value >> 24), (byte)(value >> 16),
                (byte)(value >>  8), (byte) value };
        else
            return new byte[]{
                (byte) value,        (byte)(value >>  8),
                (byte)(value >> 16), (byte)(value >> 24) };
    }

    public static byte[] fromLong(long value, ByteOrder order) {
        if (order == ByteOrder.BIG_ENDIAN)
            return new byte[]{
                (byte)(value >> 56), (byte)(value >> 48),
                (byte)(value >> 40), (byte)(value >> 32),
                (byte)(value >> 24), (byte)(value >> 16),
                (byte)(value >>  8), (byte) value };
        else
            return new byte[]{
                (byte) value,        (byte)(value >>  8),
                (byte)(value >> 16), (byte)(value >> 24),
                (byte)(value >> 32), (byte)(value >> 40),
                (byte)(value >> 48), (byte)(value >> 56) };
    }

    private ByteUtils() {}
}
