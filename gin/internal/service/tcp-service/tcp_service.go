package service

import (
	"bytes"
	"fmt"
	"io"
	"os"
	dbhandler "pjt/internal/db/db-handler"
	customEvent "pjt/internal/eventbus/event-define"
	"pjt/internal/infra/cache"
	"sync"

	"pjt/internal/transport/tcp/server/client"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type TCPService struct {
	Eventbus *eventbus.EventBus
	Dao      dbhandler.DBHandlerInterface
	Cache    *cache.Cache

	ClientManager *client.ClientManager

	FileTransferJobs map[uint32]map[string]*FileTransferJob // (clientId, filepath) -> file transfer session'
	muFileTransfer   sync.RWMutex
}

func NewTCPService(clientManager *client.ClientManager, dao dbhandler.DBHandlerInterface, cache *cache.Cache, eb *eventbus.EventBus) *TCPService {
	tcpService := &TCPService{
		Dao:      dao,
		Cache:    cache,
		Eventbus: eb,

		ClientManager: clientManager,
		// 파일 전송 프로토콜 1안 사용 시
		FileTransferJobs: make(map[uint32]map[string]*FileTransferJob),
	}
	return tcpService
}

func (t *TCPService) Handle0x010(clientId uint32) error {
	t.Eventbus.Publish(customEvent.NewTopic(customEvent.EventAType), customEvent.NewMessage("EVENTA").Add(map[string]interface{}{"test": fmt.Sprintf("connected client:%d", clientId)}))
	t.Eventbus.Publish(customEvent.NewTopic(customEvent.UpdateClientType),
		customEvent.NewMessage("tcp").Add(&customEvent.UpdateClientPayload{OldClientId: clientId, NewClientId: 1}))
	return nil
}

func (t *TCPService) Handle0xAA(clientId uint32) {
	t.Eventbus.Publish(customEvent.NewTopic(customEvent.EventAType), customEvent.NewMessage("EVENTA").Add(map[string]interface{}{"test": fmt.Sprintf("disconnected client:%d", clientId)}))

	// 파일 전송 프로토콜 1안 사용시 필요
	t.muFileTransfer.Lock()
	if fileTransferSessions, exists := t.FileTransferJobs[clientId]; exists {
		for _, fileTransferSession := range fileTransferSessions {
			fileTransferSession.Close()
		}
	}
}

func (t *TCPService) readFileToBuffer(filePath string) (*bytes.Buffer, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var buffer bytes.Buffer

	_, err = io.Copy(&buffer, file)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}
