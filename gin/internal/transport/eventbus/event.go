package eventbus

type EventType int

const (
	Invalid EventType = iota
	EVENTA
	EVENTB
)

type Event interface {
	GetPayload() any
	GetType() EventType
}

type EventA struct {
	Type    EventType
	Payload PayloadA
}

type PayloadA struct {
	Name string
	Age  int
}

func (a *EventA) GetPayload() any {
	return a.Payload
}

func (a *EventA) GetType() EventType {
	return a.Type
}

type EventB struct {
	Type    EventType
	Payload PayloadB
}

type PayloadB struct {
	Args []string
	Cmd  string
}

func (b *EventB) GetPayload() any {
	return b.Payload
}

func (b *EventB) GetType() EventType {
	return b.Type
}
