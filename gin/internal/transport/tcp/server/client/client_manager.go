package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"pjt/internal/logger"
	"pjt/internal/transport/tcp/server/handler"
	"pjt/internal/transport/tcp/server/serializer"
	"sync"
	"time"
)

type ClientManager struct {
	clients map[uint32]*Client
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc

	count int
}

func NewClientManager() *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())
	cm := &ClientManager{
		clients: make(map[uint32]*Client),
		ctx:     ctx,
		cancel:  cancel,
	}
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

func (cm *ClientManager) UpdateClient(old, new uint32) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if client, exists := cm.clients[old]; exists {
		client.ClientId = new
		cm.clients[new] = client
		delete(cm.clients, old)
		logger.Printf("[TCP] Success ClientManager.UpdateClient, cilentId: %d -> %d", old, new)
	} else {
		logger.Println("[TCP] Failed to ClientManager.UpdateClient")
	}
}

func (cm *ClientManager) SendToClientNoReply(clientId uint32, opCode byte, data []byte, sendTimeout time.Duration) error {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()
	if !exists {
		return fmt.Errorf("client not found: %d", clientId)
	}

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode
	defer client.SetSequenceNum(seqNum)

	message, _ := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
	return client.SendMessage(message, sendTimeout)
}

func (cm *ClientManager) SendToClientWithReply(clientId uint32, opCode byte, data []byte,
	sendTimeout time.Duration, replyTimeout time.Duration) (*handler.ReplyMessage, error) {
	defer func() (*handler.ReplyMessage, error) {
		if r := recover(); r != nil {
			return nil, fmt.Errorf("[TCP] client wsarecv( panic ), client id: %d", clientId)
		}
		return nil, nil
	}()

	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %d", clientId)
	}

	replyCh := make(chan *handler.ReplyMessage, 1)
	if _, exists := client.ReplyCh[opCode]; exists {
		return nil, fmt.Errorf("[TCP] already exists processing message: pending reply")
	}
	client.ReplyCh[opCode] = replyCh
	defer func() {
		close(replyCh)
		delete(client.ReplyCh, opCode)
	}()

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode
	defer client.SetSequenceNum(seqNum)
	message, _ := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
	if err := client.SendMessage(message, sendTimeout); err != nil {
		return nil, err
	}

	select {
	case <-time.After(replyTimeout):
		return nil, fmt.Errorf("[TCP] reply timeout")
	// 클라이언트가 패닉이 발생하면 해당 replyCh은 닫히게 됨 -> reply가 nil이 될 수도 있음 or close of closed channel 발생
	// panic 회복 처리 필요
	case reply := <-replyCh:
		if reply.Err != nil {
			return nil, reply.Err
		}
		return reply, nil
	}
}

func (cm *ClientManager) Broadcast(msg []byte) {

}
