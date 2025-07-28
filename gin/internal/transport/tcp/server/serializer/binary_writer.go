package serializer

import "encoding/binary"

type BinaryWriter struct {
	buffer []byte
	order  binary.ByteOrder
}

func NewBinaryWriter(order binary.ByteOrder) *BinaryWriter {
	return &BinaryWriter{
		buffer: make([]byte, 0, 256),
		order:  order,
	}
}

func (w *BinaryWriter) WriteByteOne(value byte) *BinaryWriter {
	w.buffer = append(w.buffer, value)
	return w
}

func (w *BinaryWriter) WriteBytes(values []byte) *BinaryWriter {
	w.buffer = append(w.buffer, values...)
	return w
}

func (w *BinaryWriter) WriteUint16(value uint16) *BinaryWriter {
	bytes := make([]byte, 2)
	w.order.PutUint16(bytes, value)
	w.buffer = append(w.buffer, bytes...)
	return w
}

func (w *BinaryWriter) WriteUint32(value uint32) *BinaryWriter {
	bytes := make([]byte, 4)
	w.order.PutUint32(bytes, value)
	w.buffer = append(w.buffer, bytes...)
	return w
}

func (w *BinaryWriter) WriteFixedBytes(data []byte, size int) *BinaryWriter {
	fixed := make([]byte, size)
	copy(fixed, data)
	w.buffer = append(w.buffer, fixed...)
	return w
}

func (w *BinaryWriter) Bytes() []byte {
	return w.buffer
}

func (w *BinaryWriter) Reset() {
	w.buffer = w.buffer[:0]
}

// 이때, data 슬라이스는 크기는 프로토콜에 맞는 고정된 크기여야 함.
func SerializeResponse(byteOrder binary.ByteOrder, flowControl byte, opCode byte, sequence uint16, data []byte) ([]byte, error) {

	return NewBinaryWriter(byteOrder).
		WriteByteOne(flowControl).
		WriteByteOne(opCode).
		WriteUint16(sequence).
		WriteUint16(uint16(len(data))).
		WriteBytes(data).
		Bytes(), nil
}

func CalculateRTMSLRC(byteOrder binary.ByteOrder, length uint32, sequence byte, unitNo byte, opCode uint16, data []byte) byte {
	var sum byte

	lengthBytes := make([]byte, 4)
	byteOrder.PutUint32(lengthBytes, length)
	for _, b := range lengthBytes {
		sum += b
	}

	sum += sequence + unitNo

	opcodeBytes := make([]byte, 2)
	byteOrder.PutUint16(opcodeBytes, opCode)
	for _, b := range opcodeBytes {
		sum += b
	}

	for _, b := range data {
		sum += b
	}

	return ^sum + 1
}
