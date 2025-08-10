package parser

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

type ParseState int

const (
	StateWaitingSTX ParseState = iota
	StateWaitingLength
	StateWaitingHeader
	StateWaitingData
	StateComplete
	StateError
)

/**
 * Byte Order: littleEndian 방식
 * LRC: Length ~ Data XOR value
 */

type Packet struct {
	STX      [2]byte
	Length   uint32
	Sequence byte
	UnitNo   byte
	OpCode   uint16
	Data     []byte
	LRC      byte
}

type RTMSParser struct {
	byteOrder binary.ByteOrder
	state     ParseState

	buffer         []byte
	bufferOffset   int
	initBufferSize int

	messageQueue   []*ParseMessage
	currentMessage *RTMSParsingContext
}

type RTMSParsingContext struct {
	stx      [2]byte
	length   uint32
	sequence uint8
	unitNo   uint8
	opCode   uint16
	data     []byte
	lrc      byte

	// 파싱 진행 상황
	expectedBytes int
	receivedBytes int
}

func NewRTMSParser(byteOrder binary.ByteOrder, bufferSize int) *RTMSParser {
	return &RTMSParser{
		byteOrder:      byteOrder,
		buffer:         make([]byte, bufferSize),
		initBufferSize: bufferSize,
		messageQueue:   make([]*ParseMessage, 0, 10),
	}
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
func (p *RTMSParser) Parse(conn net.Conn) (result *ParseMessage, err error) {

	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("[TCP] client unnormal exit( panic ), panic: %v", r)
		}
	}()

	if len(p.messageQueue) > 0 {
		msg := p.messageQueue[0]
		p.messageQueue = p.messageQueue[1:] // 큐에서 제거
		return msg, nil
	}

	readBuffer := make([]byte, 1024)
	p.OptimizeBuffer()

	for {
		err := conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		if err != nil {
			return nil, fmt.Errorf("[TCP] Failed to set read deadline: %v", err)
		}

		n, err := conn.Read(readBuffer)
		if err != nil {
			return nil, err
		}

		messages, err := p.Feed(readBuffer[:n])
		if err != nil {
			return nil, err
		}

		if len(messages) > 0 {
			firstMsg := messages[0]
			if len(messages) > 1 {
				p.messageQueue = append(p.messageQueue, messages[1:]...)
			}

			return firstMsg, nil
		}
	}
}

func (p *RTMSParser) Feed(data []byte) ([]*ParseMessage, error) {
	var messages []*ParseMessage

	if err := p.appendToBuffer(data); err != nil {
		return nil, err
	}

	for p.bufferOffset > 0 || p.state == StateComplete {
		switch p.state {
		case StateWaitingSTX:
			if !p.tryParseSTX(2) {
				return messages, nil
			}

		case StateWaitingLength:
			// length(4)
			if !p.tryParseLength(4) {
				return messages, nil // 더 많은 데이터 필요
			}

		case StateWaitingHeader:
			// Seq(1) + Unit(1) + OpCode(2)
			if !p.tryParseHeader(4) {
				return messages, nil // 더 많은 데이터 필요
			}

		case StateWaitingData:
			if !p.tryParseData(p.currentMessage.expectedBytes) {
				return messages, nil // 더 많은 데이터 필요
			}

		case StateComplete:
			msg, err := p.completeMessage()
			if err != nil {
				p.resetParser()
				return messages, err
			}
			messages = append(messages, msg)
			p.resetForNextMessage()

		case StateError:
			p.resetParser()
			return messages, fmt.Errorf("parser in error state")

		default:
			p.resetParser()
			return messages, fmt.Errorf("unknown parser state: %d", p.state)
		}
	}
	return messages, nil
}

func (p *RTMSParser) appendToBuffer(data []byte) error {
	needed := p.bufferOffset + len(data)

	// 버퍼 크기 reslice, 만약 데이터 크기를 제어하고 싶다면 아래 조건에서 필터.
	if needed > len(p.buffer) {
		newBuffer := make([]byte, needed*2)
		copy(newBuffer, p.buffer[:p.bufferOffset])
		p.buffer = newBuffer
	}

	copy(p.buffer[p.bufferOffset:], data)
	p.bufferOffset += len(data)

	return nil
}

func (p *RTMSParser) consumeBytes(count int) {
	if count >= p.bufferOffset {
		p.bufferOffset = 0
	} else {
		copy(p.buffer, p.buffer[count:p.bufferOffset])
		p.bufferOffset -= count
	}
}

func (p *RTMSParser) resetForNextMessage() {
	p.state = StateWaitingSTX
	p.currentMessage = nil
}

func (p *RTMSParser) resetParser() {
	p.state = StateWaitingSTX
	p.currentMessage = nil
	p.bufferOffset = 0
}

func (p *RTMSParser) tryParseSTX(needed int) bool {
	if p.bufferOffset < needed {
		return false
	}

	if p.buffer[0] != 0x7E || p.buffer[1] != 0x7E {
		p.consumeBytes(1)
		return true
	}

	// 새 메시지 컨텍스트 생성
	p.currentMessage = &RTMSParsingContext{
		stx: [2]byte{p.buffer[0], p.buffer[1]},
	}

	// 버퍼에서 STX 제거
	p.consumeBytes(needed)

	// 다음 상태로 전환
	p.state = StateWaitingLength

	return true
}

func (p *RTMSParser) tryParseLength(needed int) bool {
	if p.bufferOffset < needed {
		return false
	}
	length := p.byteOrder.Uint32(p.buffer[:needed])

	if length < 11 {
		p.state = StateError
		return true
	}

	p.currentMessage.length = length
	p.consumeBytes(needed)
	p.state = StateWaitingHeader
	return true
}

func (p *RTMSParser) tryParseHeader(needed int) bool {
	if p.bufferOffset < needed {
		return false
	}
	p.currentMessage.sequence = p.buffer[0]
	p.currentMessage.unitNo = p.buffer[1]
	p.currentMessage.opCode = p.byteOrder.Uint16(p.buffer[2:4])

	p.consumeBytes(needed)

	remainSize := int(p.currentMessage.length) - 10
	p.currentMessage.expectedBytes = remainSize
	p.currentMessage.receivedBytes = 0

	if remainSize > 0 {
		p.state = StateWaitingData
	} else {
		p.state = StateComplete
	}

	return true
}

func (p *RTMSParser) tryParseData(needed int) bool {
	if p.bufferOffset < needed {
		return false
	}

	if needed > 1 {
		p.currentMessage.data = make([]byte, needed-1)
		copy(p.currentMessage.data, p.buffer[:needed-1])
		p.currentMessage.lrc = p.buffer[needed-1]
	} else if needed == 1 {
		p.currentMessage.lrc = p.buffer[0]
	}

	p.consumeBytes(needed)

	p.state = StateComplete
	return true
}

func (p *RTMSParser) completeMessage() (*ParseMessage, error) {
	if p.currentMessage == nil {
		return nil, fmt.Errorf("[TCP] nil message")
	}

	msg := &ParseMessage{
		Protocol: ProtocolRTMS,
		Type:     fmt.Sprintf("rtms_0x%03X", p.currentMessage.opCode),
		Packet:   nil,
	}

	msg.Add(&Packet{
		STX:      p.currentMessage.stx,
		Length:   p.currentMessage.length,
		Sequence: p.currentMessage.sequence,
		UnitNo:   p.currentMessage.unitNo,
		OpCode:   p.currentMessage.opCode,
		Data:     p.currentMessage.data,
		LRC:      p.currentMessage.lrc,
	})

	return msg, nil
}

func (p *RTMSParser) OptimizeBuffer() {
	if len(p.buffer) == p.initBufferSize {
		return
	}

	if p.bufferOffset == 0 {
		if len(p.buffer) > p.initBufferSize {
			p.buffer = make([]byte, p.initBufferSize)
		}
	} else if len(p.buffer) >= p.initBufferSize*2 && p.bufferOffset < len(p.buffer)/4 {
		newSize := len(p.buffer) / 2
		if newSize < 8192 {
			newSize = 8192
		}

		newBuffer := make([]byte, newSize)
		copy(newBuffer, p.buffer[:p.bufferOffset])
		p.buffer = newBuffer
	}
}
