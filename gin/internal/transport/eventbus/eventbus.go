package eventbus

import (
	"context"
	"sync"
	"time"
)

/**
 * EventBus provides a communication mechanism between different parts of the application.
 * It implements a publish-subscribe pattern.
 */
type EventBus struct {
	subscribers map[MessageType][]chan *Message // key: MessageType -> value: []chan Message

	processedMsgs map[uint32]time.Time // processed message ID -> processed time
	cleanupTicker *time.Ticker         // to remove processed message id

	mu sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewEventBus() *EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	eb := &EventBus{
		subscribers:   make(map[MessageType][]chan *Message),
		processedMsgs: make(map[uint32]time.Time),
		cleanupTicker: time.NewTicker(time.Hour),
		ctx:           ctx,
		cancel:        cancel,
	}

	eb.wg.Add(1)
	go eb.cleanupOldMessages()

	return eb
}

func (b *EventBus) cleanupOldMessages() {
	defer func() {
		b.wg.Done()
		b.cleanupTicker.Stop()
	}()

	for {
		select {
		case <-b.cleanupTicker.C:
			b.mu.Lock()
			cutoff := time.Now().Add(-1 * time.Hour)
			for id, timestamp := range b.processedMsgs {
				if timestamp.Before(cutoff) {
					delete(b.processedMsgs, id)
				}
			}
			b.mu.Unlock()

		case <-b.ctx.Done():
			return
		}
	}
}

func (b *EventBus) Subscribe(topic MessageType) chan *Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *Message, 10)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

func (b *EventBus) Unsubscribe(topic MessageType, ch chan *Message) {
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

func (b *EventBus) Publish(topic MessageType, event *Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// duplicate check
	if _, exists := b.processedMsgs[event.Id]; exists {
		return
	}

	// record processed message
	b.processedMsgs[event.Id] = time.Now()

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
	b.cancel()
	b.wg.Wait()

	b.mu.Lock()
	defer b.mu.Unlock()

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
