package service

import "pjt/internal/transport/eventbus"

type TCPService struct {
	EventBus *eventbus.EventBus
}

func NewTCPService(eb *eventbus.EventBus) *TCPService {
	return &TCPService{
		EventBus: eb,
	}
}

func (t *TCPService) Handle0x010() error {
	t.EventBus.Publish(eventbus.EventAType, eventbus.NewMessage("EVENTA").Add(map[string]interface{}{"test": "sdfsd"}))
	return nil
}
