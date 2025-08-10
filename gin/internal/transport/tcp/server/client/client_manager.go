package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
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

func (cm *ClientManager) DisconnectClient(clientId uint32) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if client, exists := cm.clients[clientId]; exists {
		client.Close()
		cm.Unregister(clientId)
	}
}

func (cm *ClientManager) SendToClientNoReply(clientId uint32, opCode byte, data []byte, sendTimeout time.Duration) (response eventbus.TCPResponse) {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		response.Err = fmt.Errorf("[TCP] client not found: %d", clientId)
		return
	}

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode

	message, err := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
	if err != nil {
		response.Err = fmt.Errorf("[TCP] failed to serialize message: %w", err)
		return
	}

	if err := client.SendMessage(message, sendTimeout); err != nil {
		response.Err = err
		return
	}
	client.SetSequenceNum(seqNum)
	return
}

func (cm *ClientManager) SendToClientWithReply(clientId uint32, opCode byte, data []byte,
	sendTimeout time.Duration, replyTimeout time.Duration) (response eventbus.TCPResponse) {

	defer func() {
		if r := recover(); r != nil {
			response.Err = fmt.Errorf("[TCP] client wsarecv( panic ), client id: %d, panic: %v", clientId, r)
		}
	}()

	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		response.Err = fmt.Errorf("[TCP] client not found: %d", clientId)
		return
	}

	cm.mu.Lock()
	if _, exists := client.ReplyCh[opCode]; exists {
		cm.mu.Unlock()
		response.Err = fmt.Errorf("[TCP] already exists processing message: pending reply")
		return
	}

	replyCh := make(chan *handler.ReplyMessage, 1)
	client.ReplyCh[opCode] = replyCh
	cm.mu.Unlock()

	// 리소스 정리를 위한 defer
	defer func() {
		// replyCh가 nil이 아닌 경우에만 정리
		if replyCh != nil {
			close(replyCh)
			delete(client.ReplyCh, opCode)
		}
	}()

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode
	message, _ := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)

	if err := client.SendMessage(message, sendTimeout); err != nil {
		response.Err = err
		return
	}

	// 성공적으로 메시지를 보낸 경우에만 sequence number 업데이트
	client.SetSequenceNum(seqNum)

	if replyTimeout == 0 {
		replyTimeout = 5 * time.Second
	}

	select {
	case <-time.After(replyTimeout):
		response.Err = fmt.Errorf("[TCP] reply timeout")
		return
	case reply, ok := <-replyCh:
		if !ok {
			response.Err = fmt.Errorf("[TCP] reply channel closed")
			return
		}
		response.Data = reply.Payload
		response.Err = reply.Err
		return
	}
}

/**
 * cf)
 * 1. don't use gorutine without time.Sleep. ( use sync or time.Sleep )
 * 2. minimize scope of Rlock or Lock
 */
func (cm *ClientManager) Broadcast(msg []byte) {

}
