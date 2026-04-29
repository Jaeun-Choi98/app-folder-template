package implparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"

	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/server/model"
)

/**
* 텍스트 프로토콜 파서
**/

type ParseMessageText struct {
	Type     string
	ClientId uint32
	Packet   string
}

func (m *ParseMessageText) GetPacketId() any {
	return m.Type
}

func (m *ParseMessageText) GetPacket() any {
	return m.Packet
}

func (m *ParseMessageText) SetClientId(clientId uint32) {
	m.ClientId = clientId
}

func (m *ParseMessageText) GetClientId() uint32 {
	return m.ClientId
}

type TextParser struct{}

func NewTextParser() *TextParser {
	return &TextParser{}
}

func (p *TextParser) GetMinHeaderSize() int {
	return 1
}

func (p *TextParser) Parse(conn net.Conn) (tcpmd.ParseMsg, error) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("failed to read text message: %w", err)
		}
		return nil, err
	}

	content := strings.TrimSpace(line)

	// JSON 형태인지 확인
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		return p.parseJSONMessage(content)
	}

	// 단순 텍스트 메시지
	msg := &ParseMessageText{
		Type:   "text",
		Packet: content,
	}
	return msg, nil
}

func (p *TextParser) parseJSONMessage(content string) (*ParseMessageText, error) {
	var jsonMsg map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonMsg); err != nil {
		// JSON 파싱 실패시 단순 텍스트로 처리
		msg := &ParseMessageText{
			Type:   "text",
			Packet: content,
		}
		return msg, nil
	}

	msgType, _ := jsonMsg["type"].(string)
	if msgType == "" {
		msgType = "json"
	}
	jsonStr, _ := json.Marshal(jsonMsg)
	msg := &ParseMessageText{
		Type:   msgType,
		Packet: string(jsonStr),
	}
	return msg, nil
}
