package handler

import (
	"errors"
	"log"
	"pjt/internal/service"
	"pjt/internal/transport/tcp/server/parser"
	"sync"
)

var (
	errNotExistsMsgType = errors.New("not exists msg type")
	errInvalidLRC       = errors.New("invalid LRC value")
	errInvalidAssertion = errors.New("invalid assertion struct")
	//errBinaryRead       = errors.New("binary.Read failed")
)

type TypeHandlerInterface interface {
	handle(msg *parser.BaseMessage) error
}

type TypeHandlerFunc func(msg *parser.BaseMessage) error

func (f TypeHandlerFunc) handle(msg *parser.BaseMessage) error {
	return f(msg)
}

type HandlerManagerInterface interface {
	HandleMessage(msg *parser.BaseMessage) error
	GetSupportedTypes() []string
}

type HandlerManager struct {
	TCPService service.TCPServiceInterface
	handlers   map[string]TypeHandlerInterface
	mu         sync.RWMutex
}

func (h *HandlerManager) RegisterHandle(msgType string, handler TypeHandlerFunc) {
	h.handlers[msgType] = handler
}

func (h *HandlerManager) RegisterHandler(msgType string, handler TypeHandlerFunc) {
	h.handlers[msgType] = handler
}

func (h *HandlerManager) GetSupportedTypes() []string {
	types := make([]string, 0, len(h.handlers))
	for msgType := range h.handlers {
		types = append(types, msgType)
	}
	return types
}

func (h *HandlerManager) HandleMessage(msg *parser.BaseMessage) error {

	h.mu.RLock()
	handler, exists := h.handlers[msg.Type]
	h.mu.RUnlock()

	if !exists {
		return errNotExistsMsgType
	}

	return handler.handle(msg)
}

func NewHandlerManager(tcpServcie service.TCPServiceInterface) *HandlerManager {
	handler := &HandlerManager{
		TCPService: tcpServcie,
		handlers:   make(map[string]TypeHandlerInterface),
	}

	handler.RegisterHandle("text", func(msg *parser.BaseMessage) error {
		log.Printf("protocol: %v, msgType: %s, clientID: %s, Data: %v ", msg.Protocol, msg.Type, msg.ClientId, msg.Data)
		return nil
	})
	handler.RegisterHandle("json", func(msg *parser.BaseMessage) error {
		return nil
	})

	// 연결 상태 확인
	handler.RegisterHandle("rtms_0x010", handler.Handle0x010())
	// 접속 상태 확인
	handler.RegisterHandle("rmts_0x011", nil)

	return handler
}
