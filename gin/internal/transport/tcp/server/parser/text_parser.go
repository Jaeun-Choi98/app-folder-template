package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
)

/**
* 텍스트 프로토콜 파서
**/
type TextParser struct{}

func NewTextParser() *TextParser {
	return &TextParser{}
}

func (p *TextParser) GetProtocolType() ProtocolType {
	return ProtocolText
}

func (p *TextParser) CanParse(data []byte) bool {
	// 첫 몇 바이트가 ASCII 텍스트인지 확인
	if len(data) < 1 {
		return false
	}

	// 인쇄 가능한 ASCII 문자인지 확인
	for i := 0; i < len(data) && i < 4; i++ {
		if data[i] < 32 || data[i] > 126 {
			return false
		}
	}
	return true
}

func (p *TextParser) GetMinHeaderSize() int {
	return 1 // 최소 1바이트
}

func (p *TextParser) Parse(conn net.Conn) (*ParseMessage, error) {
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
	msg := &ParseMessage{
		Protocol:   ProtocolText,
		Type:       "text",
		TextPacket: content,
	}
	return msg, nil
}

func (p *TextParser) parseJSONMessage(content string) (*ParseMessage, error) {
	var jsonMsg map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonMsg); err != nil {
		// JSON 파싱 실패시 단순 텍스트로 처리
		msg := &ParseMessage{
			Protocol:   ProtocolText,
			Type:       "text",
			TextPacket: content,
		}
		return msg, nil
	}

	msgType, _ := jsonMsg["type"].(string)
	if msgType == "" {
		msgType = "json"
	}
	jsonStr, _ := json.Marshal(jsonMsg)
	msg := &ParseMessage{
		Protocol:   ProtocolText,
		Type:       msgType,
		TextPacket: string(jsonStr),
	}
	return msg, nil
}
