package parser

import (
	"errors"
	"net"
)

type ProtocolType int

const (
	ProtocolInvalid ProtocolType = iota
	ProtocolText
	ProtocolRTMS
)

var (
	errInvalidProtocolType = errors.New("invalid protocol")
	errCantParseData       = errors.New("can't parse data")
)

// 이후에 parse한 후에 msg를 handler에게 위임할 때, 필요한 context가 있다면 추가.
type BaseMessage struct {
	Protocol ProtocolType
	Type     string
	ClientId string
	Data     any
}

func (m *BaseMessage) Add(data any) *BaseMessage {
	m.Data = data
	return m
}

type Parser interface {
	GetProtocolType() ProtocolType
	CanParse(data []byte) bool
	GetMinHeaderSize() int
	Parse(conn net.Conn) (*BaseMessage, error)
}

type ParserFactory struct {
	parsers map[ProtocolType]Parser
}

func NewParserFactory() *ParserFactory {
	return &ParserFactory{
		parsers: map[ProtocolType]Parser{
			ProtocolText: NewTextParser(),
			ProtocolRTMS: NewRTMSParser(),
		},
	}
}

func (p *ParserFactory) CreateParser(protocolType ProtocolType) Parser {
	if parser, exists := p.parsers[protocolType]; exists {
		return parser
	}
	return nil
}

// 나중에 필요할
func (p *ParserFactory) DetectProtocol(data []byte) (Parser, error) {
	for _, parser := range p.parsers {
		if parser.CanParse(data) {
			return parser, nil
		}
	}
	return nil, errCantParseData
}
