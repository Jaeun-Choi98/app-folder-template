package eventbus

import (
	"time"

	"github.com/google/uuid"
)

type MessageType int

const (
	Invalid MessageType = iota
	SysSttType
	TCPNoReplyType       // 응답 없음
	TCPWithReplyType     // 응답 있음
	UpdateClientType     // 클라이언트 식별자 변경 ( 프로토콜에서 정의한 )
	DisconnectClientType // 클라이언트 세션 끊기
	EventAType
	EventBType
)

var (
	SysStt = NewMessage("sysstt")
	EventA = NewMessage("eventA")
	EventB = NewMessage("eventB")
)

type Message struct {
	Id      uint32
	Type    string
	Payload any
}

func NewMessage(t string) *Message {
	return &Message{
		Id:   uuid.New().ID(),
		Type: t,
	}
}

func (e *Message) Add(payload any) *Message {
	n := NewMessage(e.Type)
	n.Payload = payload
	return n
}

// SysStt
type SysSttPayload struct {
	ServerTime  string `json:"svrTime"`
	DBState     int8   `json:"dbComm"`
	TCPListener int8   `json:"tcpListener"`
}

// TCPClientSend
type TCPSendNoReplyPayload struct {
	ClientId    uint32
	OpCode      byte
	Data        []byte
	SendTimeout time.Duration
	Err         chan error
}

type TCPSendWithReplyPayload struct {
	ClientId     uint32
	OpCode       byte
	Data         []byte
	SendTimeout  time.Duration
	ReplyTimeout time.Duration
	Response     chan any
	Err          chan error
}

type UpdateClientPayload struct {
	OldClientId uint32
	NewClientId uint32
}

type DisconnectClientPayload struct {
	ClientId uint32
}

// Base
type BasePayload struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

func NewAddPayload() *BasePayload {
	return &BasePayload{
		Type: "add",
	}
}

func NewDeletePayload() *BasePayload {
	return &BasePayload{
		Type: "delete",
	}
}

func NewEditPayload() *BasePayload {
	return &BasePayload{
		Type: "edit",
	}
}

func (s *BasePayload) SetData(data any) *BasePayload {
	s.Data = data
	return s
}
