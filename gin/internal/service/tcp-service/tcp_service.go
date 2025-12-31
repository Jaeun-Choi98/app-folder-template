package service

import (
	"bytes"
	"fmt"
	"io"
	"os"
	dbhandler "pjt/internal/db/db-handler"
	customEvent "pjt/internal/eventbus/event-define"
	"pjt/internal/infra/cache"

	"pjt/internal/transport/tcp/server/client"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type TCPService struct {
	Eventbus       *eventbus.EventBus
	Dao            dbhandler.DBHandlerInterface
	Cache          *cache.Cache
	TCPSessionInfo *TCPSessionInfo

	ClientManager *client.ClientManager
}

/**
 * TCP 통신을 통해 받은 데이터를 저장
 */
type TCPSessionInfo struct {
	StationSession map[uint16]struct{} // 정류장 번호를 키값으로

}

func NewTCPSessionInfo() *TCPSessionInfo {
	return &TCPSessionInfo{
		StationSession: make(map[uint16]struct{}),
	}
}

func NewTCPService(clientManager *client.ClientManager, dao dbhandler.DBHandlerInterface, cache *cache.Cache, eb *eventbus.EventBus) *TCPService {
	tcpSessionInfo := NewTCPSessionInfo()
	tcpService := &TCPService{
		TCPSessionInfo: tcpSessionInfo,
		Dao:            dao,
		Cache:          cache,
		Eventbus:       eb,

		ClientManager: clientManager,
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
