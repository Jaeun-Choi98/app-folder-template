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

func (cm *ClientManager) DisconnectClient(clientId uint32) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if client, exists := cm.clients[clientId]; exists {
		client.Close()
		cm.Unregister(clientId)
	}
}

func (cm *ClientManager) SendToClientNoReply(clientId uint32, opCode byte, data []byte, sendTimeout time.Duration) error {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("[TCP] client not found: %d", clientId)
	}

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode

	message, err := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
	if err != nil {
		return fmt.Errorf("[TCP] failed to serialize message: %w", err)
	}

	if err := client.SendMessage(message, sendTimeout); err != nil {
		return err
	}
	client.SetSequenceNum(seqNum)
	return nil
}

func (cm *ClientManager) SendToClientWithReply(clientId uint32, opCode byte, data []byte,
	sendTimeout time.Duration, replyTimeout time.Duration) (result *handler.ReplyMessage, err error) {

	// panic 복구를 위한 defer - named return 사용
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("[TCP] client wsarecv( panic ), client id: %d, panic: %v", clientId, r)
		}
	}()

	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %d", clientId)
	}

	// 이미 처리 중인 요청이 있는지 확인
	if _, exists := client.ReplyCh[opCode]; exists {
		return nil, fmt.Errorf("[TCP] already exists processing message: pending reply")
	}

	replyCh := make(chan *handler.ReplyMessage, 1)
	client.ReplyCh[opCode] = replyCh

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
		return nil, err
	}

	// 성공적으로 메시지를 보낸 경우에만 sequence number 업데이트
	client.SetSequenceNum(seqNum)

	if replyTimeout == 0 {
		replyTimeout = 5 * time.Second
	}

	select {
	case <-time.After(replyTimeout):
		return nil, fmt.Errorf("[TCP] reply timeout")
	case reply, ok := <-replyCh:
		if !ok {
			// 채널이 닫힌 경우 (클라이언트 연결 끊김 등)
			return nil, fmt.Errorf("[TCP] reply channel closed")
		}
		if reply == nil {
			return nil, fmt.Errorf("[TCP] received nil reply")
		}
		return reply, reply.Err
	}
}

/**
 * cf)
 * 1. don't use gorutine without time.Sleep. ( use sync or time.Sleep )
 * 2. minimize scope of Rlock or Lock
 */
func (cm *ClientManager) Broadcast(msg []byte) {

}
