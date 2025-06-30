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

func (cm *ClientManager) Unregister(clientId string) {
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

func (cm *ClientManager) GetClient(clientId string) *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if client, exists := cm.clients[clientId]; exists {
		return client
	}
	return nil
}

// 수정 필요
func (cm *ClientManager) sendWorker(sendQueue chan *eventbus.Event) {
	for {
		select {
		case <-cm.ctx.Done():
			return
		case event := <-sendQueue:
			req, ok := event.Payload.(*eventbus.TCPClientSendPayload)
			if !ok {
				logger.Println("[TCP] ClientManager.sendWork failed to assert struct")
				continue
			}
			go func() {
				err := cm.sendToClientDirect(req.ClientId, req.Message)
				if err != nil {
					logger.Println(err)
				}
				req.Res <- err
			}()
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

func (cm *ClientManager) Broadcast(msg []byte) {

}
