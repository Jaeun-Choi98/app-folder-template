package eventbus

import (
	"sync"
)

/**
 * EventBus provides a communication mechanism between different parts of the application.
 * It implements a publish-subscribe pattern.
 */
type EventBus struct {
	subscribers map[EventType][]chan Event // key: EventType -> value: []chan Event
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan Event),
	}
}

func (b *EventBus) Subscribe(topic EventType) chan Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Event, 10)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

func (b *EventBus) Unsubscribe(topic EventType, ch chan Event) {
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

func (b *EventBus) Publish(topic EventType, event Event) {
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

func (b *EventBus) Close() {
	for topic, subs := range b.subscribers {
		for _, ch := range subs {
			for len(ch) > 0 {
				<-ch
			}
			close(ch)
		}
		delete(b.subscribers, topic)
	}
}
