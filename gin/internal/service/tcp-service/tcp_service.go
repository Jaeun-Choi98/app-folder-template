package service

import (
	"context"
	"fmt"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/ram"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/tcp/server/client"
	"sync"
)

type TCPService struct {
	Eventbus       *eventbus.EventBus
	Dao            dbhandler.DBHandlerInterface
	Ram            *ram.Ram
	TCPSessionInfo *TCPSessionInfo

	ClientManager *client.ClientManager
	NoReplyCh     chan *eventbus.Message
	WithReplyCh   chan *eventbus.Message
	UpdateCh      chan *eventbus.Message

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
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

func NewTCPService(clientManager *client.ClientManager, dao dbhandler.DBHandlerInterface, ram *ram.Ram, eb *eventbus.EventBus) *TCPService {
	tcpSessionInfo := NewTCPSessionInfo()

	noReplyCh := eb.Subscribe(eventbus.TCPNoReplyType)
	withReplyCh := eb.Subscribe(eventbus.TCPWithReplyType)
	updateCh := eb.Subscribe(eventbus.UpdateClientType)

	ctx, cancel := context.WithCancel(context.Background())
	tcpService := &TCPService{
		TCPSessionInfo: tcpSessionInfo,
		Dao:            dao,
		Ram:            ram,
		Eventbus:       eb,

		ClientManager: clientManager,
		NoReplyCh:     noReplyCh,
		WithReplyCh:   withReplyCh,
		UpdateCh:      updateCh,

		ctx:    ctx,
		cancel: cancel,
	}
	tcpService.wg.Add(1)
	go tcpService.Worker()
	return tcpService
}

func (t *TCPService) Close() {
	t.cancel()
	t.Eventbus.Unsubscribe(eventbus.TCPNoReplyType, t.NoReplyCh)
	t.Eventbus.Unsubscribe(eventbus.TCPWithReplyType, t.WithReplyCh)
	t.Eventbus.Unsubscribe(eventbus.UpdateClientType, t.UpdateCh)
	t.wg.Wait()
	logger.Println("[TCP Service] TCP Service worker goroutine is terminated")
}

func (t *TCPService) Worker() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		case msg := <-t.NoReplyCh:
			req, ok := msg.Payload.(*eventbus.TCPSendNoReplyPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct")
				continue
			}
			go func() {
				err := t.ClientManager.SendToClientNoReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout)
				if err != nil {
					logger.Println(err)
				}
				req.Err <- err
			}()
		case msg := <-t.WithReplyCh:
			req, ok := msg.Payload.(*eventbus.TCPSendWithReplyPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct")
				continue
			}
			go func() {
				res, err := t.ClientManager.SendToClientWithReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout, req.ReplyTimeout)
				if err != nil {
					logger.Printf("[TCP Service] Worker: failed to sendToClientWithReply: \n\t%v", err)
				}
				req.Err <- err
				req.Response <- res
			}()
		case msg := <-t.UpdateCh:
			req, ok := msg.Payload.(*eventbus.UpdateClientPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct")
				continue
			}
			t.ClientManager.UpdateClient(req.OldClientId, req.NewClientId)
		}
	}
}

func (t *TCPService) Handle0x010(clientId uint32) error {
	t.Eventbus.Publish(eventbus.EventAType, eventbus.NewMessage("EVENTA").Add(map[string]interface{}{"test": fmt.Sprintf("connected client:%d", clientId)}))
	return nil
}
func (t *TCPService) Handle0xAA(clientId uint32) {
	t.Eventbus.Publish(eventbus.EventAType, eventbus.NewMessage("EVENTA").Add(map[string]interface{}{"test": fmt.Sprintf("disconnected client:%d", clientId)}))
}
