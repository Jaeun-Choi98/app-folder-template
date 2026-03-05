package worker

import (
	"context"
	dbhandler "pjt/internal/db/db-handler"
	"pjt/internal/infra/cache"
	"pjt/internal/logger"
	"pjt/internal/transport/tcp/server/client"
	"sync"

	customEvent "pjt/internal/eventbus/event-define"

	"github.com/Jaeun-Choi98/modules/eventbus"
)

type TCPWorker struct {
	Eventbus *eventbus.EventBus
	Dao      dbhandler.DBHandlerInterface
	Cache    *cache.Cache

	ClientManager *client.ClientManager
	NoReplyCh     chan eventbus.Event
	WithReplyCh   chan eventbus.Event
	UpdateCh      chan eventbus.Event
	DisconnectCh  chan eventbus.Event

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewTCPWorker(d dbhandler.DBHandlerInterface, c *cache.Cache, eb *eventbus.EventBus,
	clientManager *client.ClientManager) *TCPWorker {
	noReplyCh := eb.Subscribe(customEvent.NewTopic(customEvent.TCPNoReplyType), customEvent.ChannelMore)
	withReplyCh := eb.Subscribe(customEvent.NewTopic(customEvent.TCPWithReplyType), customEvent.ChannelMore)
	updateCh := eb.Subscribe(customEvent.NewTopic(customEvent.UpdateClientType), customEvent.ChannelMore)
	disconnectCh := eb.Subscribe(customEvent.NewTopic(customEvent.DisconnectClientType), customEvent.ChannelMore)

	ctx, cancel := context.WithCancel(context.Background())
	return &TCPWorker{
		Eventbus: eb,
		Dao:      d,
		Cache:    c,

		ClientManager: clientManager,
		NoReplyCh:     noReplyCh,
		WithReplyCh:   withReplyCh,
		UpdateCh:      updateCh,
		DisconnectCh:  disconnectCh,

		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *TCPWorker) Start() {
	t.wg.Add(1)
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		case msg := <-t.NoReplyCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.TCPSendPayload)
			if !ok {
				logger.Infoln("[TCP Worker] Failed to assert struct( *customEvent.TCPSendNoReplyPayload )")
				continue
			}
			go func() {
				err := t.ClientManager.SendToClientNoReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout)
				if err != nil {
					logger.Infof("[TCP Worker] Failed to SendToClientNoReply: \n%v, station num: %d", err, req.ClientId)
				}
				req.Response <- customEvent.TCPResponse{Err: err}
			}()
		case msg := <-t.WithReplyCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.TCPSendPayload)
			if !ok {
				logger.Infoln("[TCP Worker] Failed to assert struct( *customEvent.TCPSendWithReplyPayload )")
				continue
			}
			go func() {
				reply, err := t.ClientManager.SendToClientWithReply(req.ClientId, req.OpCode, req.Data, req.SendTimeout, req.ReplyTimeout)
				if err != nil {
					logger.Infof("[TCP Worker] Failed to sendToClientWithReply: \n%v, station num: %d", err, req.ClientId)
				}
				req.Response <- customEvent.TCPResponse{Data: reply, Err: err}
			}()
		case msg := <-t.UpdateCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.UpdateClientPayload)
			if !ok {
				logger.Infoln("[TCP Worker] Failed to assert struct")
				continue
			}
			t.ClientManager.UpdateClient(req.OldClientId, req.NewClientId)
		case msg := <-t.DisconnectCh:
			req, ok := msg.(*customEvent.Message).Payload.(*customEvent.DisconnectClientPayload)
			if !ok {
				logger.Infoln("[TCP Worker] Failed to assert struct")
				continue
			}
			t.ClientManager.DisconnectClient(req.ClientId)
		}
	}
}

func (t *TCPWorker) Shutdown() {
	t.cancel()
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.TCPNoReplyType), t.NoReplyCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.TCPWithReplyType), t.WithReplyCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.UpdateClientType), t.UpdateCh)
	t.Eventbus.Unsubscribe(customEvent.NewTopic(customEvent.DisconnectClientType), t.DisconnectCh)
	t.wg.Wait()
	logger.Infoln("[TCP Worker] TCP Worker goroutine is terminated")
}
