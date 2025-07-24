package client

import (
	"context"
	"fmt"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/tcp/server/handler"
	"sync"
	"time"
)

type ClientManager struct {
	clients  map[uint32]*Client
	eventbus *eventbus.EventBus
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc

	count int
}

func NewClientManager(pctx context.Context, eb *eventbus.EventBus) *ClientManager {
	ctx, cancel := context.WithCancel(pctx)
	cm := &ClientManager{
		clients:  make(map[uint32]*Client),
		eventbus: eb,
		ctx:      ctx,
		cancel:   cancel,
	}
	sendQueue1 := eb.Subscribe(eventbus.TCPNoReplyType)
	sendQueue2 := eb.Subscribe(eventbus.TCPWithReplyType)
	updateQueue := eb.Subscribe(eventbus.UpdateClientType)
	go cm.Worker(sendQueue1, sendQueue2, updateQueue)
	return cm
}

func (cm *ClientManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cancel()
	for _, client := range cm.clients {
		client.Close()
	}
	cm.count = 0
}

func (cm *ClientManager) Register(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.clients[client.ClientId]; !exists {
		cm.clients[client.ClientId] = client
		cm.count++
	}
}

func (cm *ClientManager) Unregister(clientId uint32) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, exists := cm.clients[clientId]; exists {
		delete(cm.clients, clientId)
		cm.count--
	}
}

func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.count
}

func (cm *ClientManager) GetClient(clientId uint32) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if client, exists := cm.clients[clientId]; exists {
		return client
	}
	return nil
}

func (cm *ClientManager) Worker(sendQueue1, sendQueue2, updateQueue chan *eventbus.Message) {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case msg := <-sendQueue1:
			req, ok := msg.Payload.(*eventbus.TCPSendNoReplyPayload)
			if !ok {
				logger.Println("[TCP] ClientManager.worker failed to assert struct")
				continue
			}
			go func() {
				err := cm.sendToClientNoReply(req.ClientId, req.Data, req.SendTimeout)
				if err != nil {
					logger.Println(err)
				}
				req.Err <- err
			}()
		case msg := <-sendQueue2:
			req, ok := msg.Payload.(*eventbus.TCPSendWithReplyPayload)
			if !ok {
				logger.Println("[TCP] ClientManager.worker failed to assert struct")
				continue
			}
			go func() {
				res, err := cm.sendToClientWithReply(req.ClientId, req.Data, req.SendTimeout, req.ReplyTimeout)
				if err != nil {
					logger.Println(err)
				}
				req.Err <- err
				req.Response <- res
			}()
		case msg := <-updateQueue:
			logger.Printf("%v ", msg)
		}
	}
}

func (cm *ClientManager) sendToClientNoReply(clientId uint32, message []byte, sendTimeout time.Duration) error {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %d", clientId)
	}

	return client.SendMessage(message, sendTimeout)
}

func (cm *ClientManager) sendToClientWithReply(clientId uint32, message []byte,
	sendTimeout time.Duration, replyTimeout time.Duration) (*handler.ReplyMessage, error) {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %d", clientId)
	}
	replyCh := make(chan *handler.ReplyMessage, 1)

	// Message의 OPCODE를 Key로 줄 생각.
	if _, exists := client.ReplyCh[0x02]; exists {
		return nil, fmt.Errorf("[TCP] already exists processing message: pending message")
	}
	client.ReplyCh[0x02] = replyCh
	defer func() {
		close(replyCh)
		delete(client.ReplyCh, 0x02)
	}()

	if err := client.SendMessage(message, sendTimeout); err != nil {
		return nil, err
	}

	select {
	case <-time.After(replyTimeout):
		return nil, fmt.Errorf("[TCP] reply timeout")
	case reply := <-replyCh:
		if reply.Err != nil {
			return nil, reply.Err
		}
		return reply, nil
	}
}

func (cm *ClientManager) Broadcast(msg []byte) {

}
