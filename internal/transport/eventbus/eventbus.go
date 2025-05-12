package eventbus

import (
	"sync"

	model "root/internal/model/event"
)

/**
 * EventBus provides a communication mechanism between different parts of the application.
 * It implements a publish-subscribe pattern.
 */
type EventBus struct {
	subscribers map[string][]chan model.Event
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan model.Event),
	}
}

// Subscribe returns a channel that will receive events published to the given topic
func (b *EventBus) Subscribe(topic string) chan model.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan model.Event, 1)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

// Unsubscribe removes a subscription
func (b *EventBus) Unsubscribe(topic string, ch chan model.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, found := b.subscribers[topic]; found {
		for i, sub := range subs {
			if sub == ch {
				close(ch)
				b.subscribers[topic] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}
}

// Publish sends an event to all subscribers of the given topic
func (b *EventBus) Publish(topic string, event model.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, found := b.subscribers[topic]; found {
		for _, ch := range subs {
			// Use non-blocking send to avoid deadlocks
			select {
			case ch <- event:
			default:
				// Channel is full, log this if needed
			}
		}
	}
}
