package handler

import (
	"errors"
	"log"
	"pjt/internal/logger"
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

type ReplyMessage struct {
	Payload any
	Err     error
}

type TypeHandlerInterface interface {
	handle(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error
}

type TypeHandlerFunc func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error

func (f TypeHandlerFunc) handle(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
	return f(msg, replyCh)
}

type HandlerManagerInterface interface {
	HandleMessage(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error
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

func (h *HandlerManager) HandleMessage(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {

	h.mu.RLock()
	handler, exists := h.handlers[msg.Type]
	h.mu.RUnlock()

	if !exists {
		return errNotExistsMsgType
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Println("[TCP] Panic in handling message")
			}
		}()
		if err := handler.handle(msg, replyCh); err != nil {
			logger.Printf("[TCP] handler error for %s (client %d): %v",
				msg.Type, msg.ClientId, err)
		}
	}()
	return nil
}

func NewHandlerManager(tcpServcie service.TCPServiceInterface) *HandlerManager {
	handler := &HandlerManager{
		TCPService: tcpServcie,
		handlers:   make(map[string]TypeHandlerInterface),
	}

	handler.RegisterHandle("text", func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		log.Printf("protocol: %v, msgType: %s, clientID: %d, Data: %v ", msg.Protocol, msg.Type, msg.ClientId, msg.TextPacket)
		return nil
	})
	handler.RegisterHandle("json", func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		return nil
	})

	handler.RegisterHandle("rtms_0x010", handler.Handle0x010())
	// 연결 상태 확인
	handler.RegisterHandle("rmts_0x001", handler.Handle0x001())
	// reply test
	handler.RegisterHandle("rtms_0x002", handler.Handle0x002())

	// close client
	handler.RegisterHandle("rtms_0x0AA", handler.Handle0xAA())
	return handler
}
