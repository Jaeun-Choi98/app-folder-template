package service

import (
	"bytes"
	"io"
	"os"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/event"
	"pjt/internal/infra/cache"
	"sync"

	"pjt/internal/transport/tcp/server/client"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type TCPService struct {
	Dao   dbhandler.DBHandlerInterface
	Cache *cache.Cache

	ClientManager *client.ClientManager

	FileTransferJobs map[uint32]map[string]*FileTransferJob // (clientId, filepath) -> file transfer session'
	muFileTransfer   sync.RWMutex
}

func NewTCPService(clientManager *client.ClientManager, dao dbhandler.DBHandlerInterface, cache *cache.Cache) *TCPService {
	tcpService := &TCPService{
		Dao:   dao,
		Cache: cache,

		ClientManager: clientManager,
		// 파일 전송 프로토콜 1안 사용 시
		FileTransferJobs: make(map[uint32]map[string]*FileTransferJob),
	}
	return tcpService
}

func (t *TCPService) Handle0x010(clientId uint32) error {
	eventbus.Publish(event.GetEventBus(), "sse.event", &event.SseEvent{Event: "tcp service", Data: "hello world"})
	t.ClientManager.UpdateClient(clientId, 1)
	return nil
}

func (t *TCPService) Handle0xAA(clientId uint32) {

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
