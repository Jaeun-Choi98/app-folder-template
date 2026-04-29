package event

import (
	"time"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

var eb *eventbus.EventBus

func SetEventBus(e *eventbus.EventBus) {
	eb = e
}

func GetEventBus() *eventbus.EventBus {
	return eb
}

type SseEvent struct {
	Event string
	Data  any
}

/**
 * TcpEventOpcode
 * 0x01: 응답 없는 메시지 전송
 * 0x02: 응답 있는 메시지 전송
 * 0x03: 클라이언트 아이디 업데이트
 * 0x04: 클라리언트 연결 끊음
 */
type TcpEvent struct {
	TcpEventOpcode byte
	ClientId       uint32
	PacketOpcode   byte
	Data           []byte
	SendTimeout    time.Duration
	ReplyTimeout   time.Duration
}
