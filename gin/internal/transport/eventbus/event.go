package eventbus

import "time"

type EventType int

const (
	Invalid EventType = iota
	SysSttType
	TCPSendType
	EventAType
	EventBType
)

var (
	SysStt = NewEvent("sysstt")
	EventA = NewEvent("eventA")
	EventB = NewEvent("eventB")
)

type Event struct {
	Type    string
	Payload any
}

func NewEvent(t string) *Event {
	return &Event{
		Type: t,
	}
}

func (e *Event) Add(payload any) *Event {
	n := NewEvent(e.Type)
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
type TCPClientSendPayload struct {
	ClientId string
	Message  []byte
	Timeout  time.Duration
	Res      chan error
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
