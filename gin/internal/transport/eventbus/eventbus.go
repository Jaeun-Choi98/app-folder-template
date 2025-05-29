package eventbus

import (
	"sync"

	model "pjt/internal/model/event"
)

/**
 * EventBus provides a communication mechanism between different parts of the application.
 * It implements a publish-subscribe pattern.
 */
type EventBus struct {
	subscribers map[model.EventType][]chan model.Event // key: EventType -> value: []chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[model.EventType][]chan model.Event),
	}
}

func (b *EventBus) Subscribe(topic model.EventType) chan model.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan model.Event, 10)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

func (b *EventBus) Unsubscribe(topic model.EventType, ch chan model.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, exsits := b.subscribers[topic]; exsits {
		for i, sub := range subs {
			if sub == ch {
				close(ch)
				b.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

func (b *EventBus) Publish(topic model.EventType, event model.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, found := b.subscribers[topic]; found {
		for _, ch := range subs {
			select {
			case ch <- event:
			default:

			}
		}
	}
}
