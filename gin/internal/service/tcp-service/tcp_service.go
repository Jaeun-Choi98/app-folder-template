package service

import (
	"context"
	"fmt"
	dbhandler "pjt/internal/db/db-handler"
	customEvent "pjt/internal/eventbus/event-define"
	"pjt/internal/infra/cache"
	"pjt/internal/logger"

	"pjt/internal/transport/tcp/server/client"
	"sync"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type TCPService struct {
	Eventbus       *eventbus.EventBus
	Dao            dbhandler.DBHandlerInterface
	Cache          *cache.Cache
	TCPSessionInfo *TCPSessionInfo

	ClientManager *client.ClientManager
	NoReplyCh     chan eventbus.Event
	WithReplyCh   chan eventbus.Event
	UpdateCh      chan eventbus.Event
	DisconnectCh  chan eventbus.Event

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

func NewTCPService(clientManager *client.ClientManager, dao dbhandler.DBHandlerInterface, cache *cache.Cache, eb *eventbus.EventBus) *TCPService {
	tcpSessionInfo := NewTCPSessionInfo()

	noReplyCh := eb.Subscribe(customEvent.NewTopic(customEvent.TCPNoReplyType), customEvent.ChannelMore)
	withReplyCh := eb.Subscribe(customEvent.NewTopic(customEvent.TCPWithReplyType), customEvent.ChannelMore)
	updateCh := eb.Subscribe(customEvent.NewTopic(customEvent.UpdateClientType), customEvent.ChannelMore)
	disconnectCh := eb.Subscribe(customEvent.NewTopic(customEvent.DisconnectClientType), customEvent.ChannelMore)

	ctx, cancel := context.WithCancel(context.Background())
	tcpService := &TCPService{
		TCPSessionInfo: tcpSessionInfo,
		Dao:            dao,
		Cache:          cache,
		Eventbus:       eb,

		ClientManager: clientManager,
		NoReplyCh:     noReplyCh,
		WithReplyCh:   withReplyCh,
		UpdateCh:      updateCh,
		DisconnectCh:  disconnectCh,

		ctx:    ctx,
		cancel: cancel,
	}
	tcpService.wg.Add(1)
	go tcpService.Worker()
	return tcpService
}

func (t *TCPService) Close() {
	t.cancel()
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.TCPNoReplyType), t.NoReplyCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.TCPWithReplyType), t.WithReplyCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.UpdateClientType), t.UpdateCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.DisconnectClientType), t.DisconnectCh)
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
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.TCPSendPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct( *customEvent.TCPSendNoReplyPayload )")
				continue
			}
			go func() {
				res := t.ClientManager.SendToClientNoReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout)
				if res.Err != nil {
					logger.Printf("[TCP Service] Worker: failed to SendToClientNoReply: \n%v, station num: %d", res.Err, req.ClientId)
				}
				req.Response <- res
			}()
		case msg := <-t.WithReplyCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.TCPSendPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct( *customEvent.TCPSendWithReplyPayload )")
				continue
			}
			go func() {
				res := t.ClientManager.SendToClientWithReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout, req.ReplyTimeout)
				if res.Err != nil {
					logger.Printf("[TCP Service] Worker: failed to sendToClientWithReply: \n%v, station num: %d", res.Err, req.ClientId)
				}
				req.Response <- res
			}()
		case msg := <-t.UpdateCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.UpdateClientPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct")
				continue
			}
			t.ClientManager.UpdateClient(req.OldClientId, req.NewClientId)
		case msg := <-t.DisconnectCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.DisconnectClientPayload)
			if !ok {
				logger.Println("[TCP Service] Worker: failed to assert struct")
				continue
			}
			t.ClientManager.DisconnectClient(req.ClientId)
		}
	}
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
