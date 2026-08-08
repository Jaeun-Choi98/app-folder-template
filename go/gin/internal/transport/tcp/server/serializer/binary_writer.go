package serializer

import (
	"encoding/binary"

	"github.com/Jaeun-Choi98/modules/tcpnet/advanced/serializer"
)

// 이때, data 슬라이스는 크기는 프로토콜에 맞는 고정된 크기여야 함.
func SerializeResponse(byteOrder binary.ByteOrder, flowControl byte, opCode byte, sequence uint16, data []byte) ([]byte, error) {

	return serializer.NewBinaryWriter(byteOrder).
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
