package client

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"pjt/internal/logger"

	"pjt/internal/transport/tcp/server/serializer"
	"sync"
	"time"

	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/server/model"
)

// for distingush
var (
	ErrNormal       = fmt.Errorf("normal error")
	ErrTimeoutSend  = fmt.Errorf("tcp send timeout")
	ErrTimeoutReply = fmt.Errorf("tcp reply timeout")
)

type ClientManager struct {
	clients map[uint32]*Client
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc

	count int

	// for send logical packet
	sendMu sync.RWMutex
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
		client.IsAuth = true
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

func (cm *ClientManager) SendToClientNoReply(clientId uint32, opCode byte, data []byte, sendTimeout time.Duration) (retErr error) {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		logger.Printf("[TCP] client not found: %d, in processing op_code: %+v", clientId, opCode)
		retErr = ErrNormal
		return
	}

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode

	message, err := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
	if err != nil {
		logger.Println("[TCP] failed to serialize message")
		retErr = ErrNormal
		return
	}

	cm.sendMu.Lock()
	if err := client.SendMessage(message, sendTimeout); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			retErr = ErrTimeoutSend
		} else {
			retErr = err
		}
		logger.Printf("[TCP] failed to sendmessage, err: %+v", err)
		return
	}
	cm.sendMu.Unlock()
	logger.Printf("[TCP] Transmitted packet, client id: %+v, Flow Control: %+v, OP Code: 0x%02X, SeqNum: %+v, Data Length: %+v, Data: [%x]",
		clientId, 0x00, opCode, seqNum, len(data), data)

	client.SetSequenceNum(seqNum)
	return
}

func (cm *ClientManager) SendToClientWithReply(clientId uint32, opCode byte, data []byte,
	sendTimeout time.Duration, replyTimeout time.Duration) (replyData any, retErr error) {

	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("[TCP] client wsarecv( panic ), client id: %d, panic: %v", clientId, r)
		}
	}()

	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		logger.Printf("[TCP] client not found: %d, in processing op_code: %+v", clientId, opCode)
		retErr = ErrNormal
		return
	}

	if _, exists := client.clientContext.GetReplyChannel().Get(opCode); !exists {
		replyCh := make(chan tcpmd.Reply, 1)
		client.clientContext.GetReplyChannel().Set(opCode, replyCh)
	}

	replyCh, _ := client.clientContext.GetReplyChannel().Get(opCode)

	seqNum := (client.GetSequenceNum() + 1) % SequenceMode
	message, _ := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)

	cm.sendMu.Lock()
	if err := client.SendMessage(message, sendTimeout); err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			retErr = ErrTimeoutSend
		} else {
			retErr = err
		}
		logger.Printf("[TCP] failed to sendmessage, err: %+v", err)
		return
	}
	cm.sendMu.Unlock()

	logger.Printf("[TCP] Transmitted packet, client id: %+v, Flow Control: %+v, OP Code: 0x%02X, SeqNum: %+v, Data Length: %+v, Data: [%x]",
		clientId, 0x00, opCode, seqNum, len(data), data)

	client.SetSequenceNum(seqNum)

	if replyTimeout == 0 {
		replyTimeout = 5 * time.Second
	}

	select {
	case <-time.After(replyTimeout):
		logger.Printf("[TCP] %+v reply timeout, client id: %+v", opCode, clientId)
		retErr = ErrTimeoutReply
		return
	case reply, ok := <-replyCh:
		if !ok {
			logger.Printf("[TCP] %+v reply channel closed, client id: %+v", opCode, clientId)
			retErr = ErrNormal
			return
		}
		replyData = reply.GetPayload()
		retErr = reply.GetErr()
		return
	}
}

func (cm *ClientManager) SendFileToClient(clientId uint32, opCode byte, buf []byte, sendTimeout time.Duration) (retErr error) {
	cm.mu.RLock()
	client, exists := cm.clients[clientId]
	cm.mu.RUnlock()

	if !exists {
		logger.Printf("[TCP] client not found: %d, in processing op_code: 0x%02X", clientId, opCode)
		retErr = ErrNormal
		return
	}

	offset := 0
	maxChunk := 1 * 1024 * 1024
	totalSize := len(buf)
	var chunkCnt, chunkIdx int
	a, b := totalSize/maxChunk, totalSize%maxChunk
	if b > 0 {
		chunkCnt = a + 1
	} else {
		chunkCnt = a
	}

	// 청크 인덱스는 1부터
	chunkIdx = 1

	var seqNums []uint16
	for offset < len(buf) {
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, clientId)
		data = append(data, byte(chunkCnt), byte(chunkIdx))
		if offset+maxChunk > len(buf) {
			chunkSise := make([]byte, 4)
			binary.BigEndian.PutUint32(chunkSise, uint32(totalSize-offset))
			data = append(data, chunkSise...)
			data = append(data, buf[offset:]...)
			offset += len(buf)
		} else {
			chunkSise := make([]byte, 4)
			binary.BigEndian.PutUint32(chunkSise, uint32(maxChunk))
			data = append(data, chunkSise...)
			data = append(data, buf[offset:offset+maxChunk]...)
			offset += maxChunk
		}
		chunkIdx += 1

		seqNum := (client.GetSequenceNum() + 1) % SequenceMode

		message, err := serializer.SerializeResponse(binary.BigEndian, 0x00, opCode, seqNum, data)
		if err != nil {
			logger.Println("[TCP] failed to serialize message")
			retErr = ErrNormal
			return
		}

		cm.sendMu.Lock()
		if err := client.SendMessage(message, sendTimeout); err != nil {
			var sendErr error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				sendErr = ErrTimeoutSend
			} else {
				sendErr = err
			}
			logger.Printf("[TCP] failed to send message, err: %+v", err)
			retErr = sendErr
			return
		}
		cm.sendMu.Unlock()

		client.SetSequenceNum(seqNum)
		seqNums = append(seqNums, seqNum)
	}

	logger.Printf("[TCP] Transmitted packet, client id: %+v, Flow Control: %+v, OP Code: 0x%02X, SeqNum: %+v, Data Length: %+v, Chunk count: %v, ",
		clientId, 0x00, opCode, seqNums, len(buf), chunkCnt)

	return
}

/**
 * cf)
 * 1. don't use gorutine without time.Sleep. ( use sync or time.Sleep )
 * 2. minimize scope of Rlock or Lock
 */
func (cm *ClientManager) Broadcast(msg []byte, opCode any) {
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("[TCP] Failed to broadcast( panic ), OPCODE: %+v, panic: %v", opCode, r)
		}
	}()
}
