package eventbus

type EventType int

const (
	Invalid EventType = iota
	SysSttType
	EVENTA
	EVENTB
)

type Event interface {
	GetPayload() any
	GetType() string
}

type SysStt struct {
	Type    string
	Payload any
}

func (s *SysStt) GetPayload() any {
	return s.Payload
}

func (s *SysStt) GetType() string {
	return s.Type
}

type EventA struct {
	Type    string
	Payload PayloadA
}

type PayloadA struct {
	Name string
	Age  int
}

func (a *EventA) GetPayload() any {
	return a.Payload
}

func (a *EventA) GetType() string {
	return a.Type
}

type EventB struct {
	Type    string
	Payload PayloadB
}

type PayloadB struct {
	Args []string
	Cmd  string
}

func (b *EventB) GetPayload() any {
	return b.Payload
}

func (b *EventB) GetType() string {
	return b.Type
}
