package parser

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

/**
 * Byte Order: littleEndian 방식
 * LRC: Length ~ Data XOR value
 */

type RTMSMessage struct {
	STX      [2]byte
	Length   uint32
	Sequence byte
	UnitNo   byte
	OpCode   uint16
	Data     []byte
	LRC      byte
}

func (m *RTMSMessage) ToByte() []byte {
	b := make([]byte, m.Length)
	b = append(b, m.LRC)
	b = append(b, m.Data...)
	b = binary.LittleEndian.AppendUint16(b, m.OpCode)
	b = append(b, m.UnitNo, m.Sequence)
	b = binary.LittleEndian.AppendUint32(b, m.Length)
	b = append(b, m.STX[1], m.STX[0])
	return b
}

type RTMSParser struct{}

func NewRTMSParser() *RTMSParser {
	return &RTMSParser{}
}

func (p *RTMSParser) GetProtocolType() ProtocolType {
	return ProtocolRTMS
}

func (p *RTMSParser) CanParse(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x7E && data[1] == 0x7E
}

// STX(2) + Length(4) + Seq(1) + Unit(1) + OpCode(2) + Data1(1)
func (p *RTMSParser) GetMinHeaderSize() int {
	return 11
}

// lrc 체크의 경우, Data값이 OPCODE에 따라 다르기 때문에 핸들러에서 처리 ( data 값이 여러 개일 수도 있음 )
func (p *RTMSParser) Parse(conn net.Conn) (*BaseMessage, error) {

	// STX 읽기
	stx := make([]byte, 2)
	if _, err := io.ReadFull(conn, stx); err != nil {
		return nil, err //fmt.Errorf("failed to read STX: %w", err)
	}

	if stx[0] != 0x7E || stx[1] != 0x7E {
		return nil, fmt.Errorf("invalid STX: %02X %02X", stx[0], stx[1])
	}

	// Length 읽기
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err //fmt.Errorf("failed to read length: %w", err)
	}

	length := binary.LittleEndian.Uint32(lengthBytes)
	if length < 11 || length > 4096 {
		return nil, fmt.Errorf("invalid length: %d", length)
	}

	// 나머지 헤더 읽기, Seq(1) + Unit(1) + OpCode(2) + Data1(1)
	headerRest := make([]byte, 4)
	if _, err := io.ReadFull(conn, headerRest); err != nil {
		return nil, err //fmt.Errorf("failed to read header: %w", err)
	}

	sequence := headerRest[0]
	unitNo := headerRest[1]
	opCode := binary.LittleEndian.Uint16(headerRest[2:4])

	// Data2와 LRC 읽기, 전체 - STX(2) - Length(4) - Header(4)
	remainingSize := int(length) - 10
	remaining := make([]byte, remainingSize)
	if _, err := io.ReadFull(conn, remaining); err != nil {
		return nil, err //fmt.Errorf("failed to read remaining data: %w", err)
	}

	// Data와 LRC 분리
	var data []byte
	var lrc byte

	if remainingSize > 1 {
		data = remaining[:remainingSize-1]
		lrc = remaining[remainingSize-1]
	} else if remainingSize == 1 {
		lrc = remaining[0]
	}

	// 메시지 생성
	msg := &BaseMessage{
		Protocol: ProtocolRTMS,
		Type:     fmt.Sprintf("rtms_0x%03X", opCode),
		Data:     nil, // 나중에 설정
	}

	msg.Add(&RTMSMessage{
		STX:      [2]byte{stx[0], stx[1]},
		Length:   length,
		Sequence: sequence,
		UnitNo:   unitNo,
		OpCode:   opCode,
		Data:     data,
		LRC:      lrc,
	})

	return msg, nil
}
