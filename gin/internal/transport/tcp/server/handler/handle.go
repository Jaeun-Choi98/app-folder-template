package handler

import (
	"errors"
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/tcp/server/parser"
	"sync"
)

var (
	errNotExistsMsgType = errors.New("not exists msg type")
)

type TypeHandler interface {
	handle(msg *parser.BaseMessage) error
}

type TypeHandlerFunc func(msg *parser.BaseMessage) error

func (f TypeHandlerFunc) handle(msg *parser.BaseMessage) error {
	return f(msg)
}

type MessageHandler struct {
	eventBus *eventbus.EventBus
	handlers map[string]TypeHandler
	mu       sync.RWMutex
}

func NewMessageHandler(eventbus *eventbus.EventBus) *MessageHandler {
	handler := &MessageHandler{
		eventBus: eventbus,
		handlers: make(map[string]TypeHandler),
	}

	handler.RegisterHandle("text", func(msg *parser.BaseMessage) error {
		return nil
	})
	handler.RegisterHandle("json", func(msg *parser.BaseMessage) error {
		return nil
	})
	handler.RegisterHandle("rtms_0x41", HandleRTMSTrainWarning(eventbus))
	return handler
}

func (h *MessageHandler) RegisterHandle(msgType string, handler TypeHandlerFunc) {
	h.handlers[msgType] = handler
}

func (h *MessageHandler) RegisterHandler(msgType string, handler TypeHandler) {
	h.handlers[msgType] = handler
}

func (h *MessageHandler) GetSupportedTypes() []string {
	types := make([]string, 0, len(h.handlers))
	for msgType := range h.handlers {
		types = append(types, msgType)
	}
	return types
}

func (h *MessageHandler) HandleMessage(msg *parser.BaseMessage) error {

	h.mu.RLock()
	handler, exists := h.handlers[msg.Type]
	h.mu.RUnlock()

	if !exists {
		return errNotExistsMsgType
	}
	handler.handle(msg)
	return nil
}
