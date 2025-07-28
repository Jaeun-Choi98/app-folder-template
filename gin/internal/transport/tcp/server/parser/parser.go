package parser

import (
	"net"
)

type ProtocolType int

const (
	ProtocolInvalid ProtocolType = iota
	ProtocolText
	ProtocolRTMS
)

var (
// errInvalidProtocolType = errors.New("invalid protocol")
// errCantParseData = errors.New("can't parse data")
)

// 이후에 parse한 후에 msg를 handler에게 위임할 때, 필요한 context가 있다면 추가.
type ParseMessage struct {
	Protocol   ProtocolType
	Type       string
	ClientId   uint32
	Packet     *Packet
	TextPacket string
}

func (m *ParseMessage) Add(data *Packet) *ParseMessage {
	m.Packet = data
	return m
}

type Parser interface {
	GetProtocolType() ProtocolType
	CanParse(data []byte) bool
	GetMinHeaderSize() int
	Parse(conn net.Conn) (*ParseMessage, error)
}
