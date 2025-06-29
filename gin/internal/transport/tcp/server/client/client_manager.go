package client

import (
	"context"
	"fmt"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"sync"
)

type ClientManager struct {
	clients  map[string]*Client
	eventbus *eventbus.EventBus
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc

	count int
}

func NewClientManager(pctx context.Context, eb *eventbus.EventBus) *ClientManager {
	ctx, cancel := context.WithCancel(pctx)
	cm := &ClientManager{
		clients:  make(map[string]*Client),
		eventbus: eb,
		ctx:      ctx,
		cancel:   cancel,
	}
	sendQueue := eb.Subscribe(eventbus.TCPSendType)
	go cm.sendWorker(sendQueue)
	return cm
}

// 비동기 전송 처리
func (c *ClientManager) sendWorker(sendQueue chan *eventbus.Event) {
	for {
		select {
		case <-c.ctx.Done():
			return
		case event := <-sendQueue:
			req := event.Payload.(*eventbus.TCPClientSendPayload)
			err := c.sendToClientDirect(req.ClientId, req.Message)
			logger.Println(err)
		}
	}
}

func (cm *ClientManager) sendToClientDirect(clientID string, message []byte) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[clientID]

	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	return client.SendMessage(message)
}

func (c *ClientManager) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cancel()
	for _, client := range c.clients {
		client.Close()
	}
	c.count = 0
}

func (c *ClientManager) Register(client *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.clients[client.ClientId]; !exists {
		c.clients[client.ClientId] = client
		c.count++
	}
}

func (c *ClientManager) Unregister(clientId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.clients[clientId]; exists {
		delete(c.clients, clientId)
		c.count--
	}
}

func (c *ClientManager) GetClientCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

func (c *ClientManager) GetClient(clientId string) *Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if client, exists := c.clients[clientId]; exists {
		return client
	}
	return nil
}

func (c *ClientManager) Broadcast(msg []byte) {

}
