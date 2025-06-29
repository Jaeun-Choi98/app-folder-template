package tcp

import (
	"context"
	"net"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/tcp/server/client"
	"pjt/internal/transport/tcp/server/handler"
	"pjt/internal/transport/tcp/server/parser"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TCPServer struct {
	listener net.Listener
	clients  *client.ClientManager

	parser  parser.Parser
	handler *handler.MessageHandler

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	isListening bool
	heartbeat   time.Duration

	timeOutCnt    int
	maxTimeOutCnt int
}

func NewTCPServer(eventBus *eventbus.EventBus, heartbeat time.Duration) (*TCPServer, error) {

	ctx, cancel := context.WithCancel(context.Background())
	parserFactory := parser.NewParserFactory()

	return &TCPServer{
		clients: client.NewClientManager(ctx, eventBus),
		parser:  parserFactory.CreateParser(parser.ProtocolText),
		handler: handler.NewMessageHandler(eventBus),
		//msgChannel:    make(chan string, 10),
		//eventBus:      eventBus,
		ctx:           ctx,
		cancel:        cancel,
		heartbeat:     heartbeat,
		maxTimeOutCnt: 3,
	}, nil
}

func (t *TCPServer) IsListening() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.isListening
}

func (t *TCPServer) SetListeningState(state bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isListening = state
	t.listener = nil
}

func (t *TCPServer) Listening() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.listener == nil {
		listener, err := net.Listen("tcp", ":5000")
		if err != nil {
			logger.Println(err)
			return err
		}
		t.listener = listener
	}

	t.isListening = true

	return nil
}

// func (t *TCPServer) registerClient(clientId string, client *client.TCPClient) {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	if _, exists := t.clients[clientId]; !exists {
// 		t.clients[clientId] = client
// 	}
// }

// func (t *TCPServer) unregisterClient(clientId string) {
// 	t.mu.Lock()
// 	defer t.mu.Unlock()

// 	delete(t.clients, clientId)
// }

func (t *TCPServer) Start() error {
	//wg.Add(1)
	//go t.SendMessageRoutine()

	if err := t.Listening(); err != nil {
		return err
	}

	t.wg.Add(1)
	return t.WaitForAccept()
}

func (t *TCPServer) WaitForAccept() error {
	defer t.wg.Done()

	for {
		t.mu.RLock()
		listener := t.listener
		isListening := t.isListening
		t.mu.RUnlock()

		select {
		case <-t.ctx.Done():
			return nil
		default:
		}

		if listener == nil || !isListening {
			time.Sleep(1 * time.Second)
			continue
		}

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.ctx.Done():
				return nil
			default:
				logger.Println(err)
				if t.IsListening() {
					t.SetListeningState(false)
				}
			}
			continue
		}
		t.wg.Add(1)
		go t.HandleConnection(t.ctx, conn)
	}
}

func (t *TCPServer) HandleConnection(ctx context.Context, conn net.Conn) {

	c := client.NewClient(t.ctx, uuid.New().String(), conn, t.parser, t.handler)
	defer func() {
		t.clients.Unregister(c.ClientId)
		c.Close()
		logger.Printf("[TCP] Disconnected client: %s", c.Conn.RemoteAddr().String())
		t.wg.Done()
	}()

	t.clients.Register(c)
	logger.Printf("[TCP] Connected new client: %s", c.Conn.RemoteAddr().String())
	c.MessageProcessingLoop()
}

// if need to send to particular client, add parameter clientId
// func (t *TCPServer) handleMessage(msg string) {
// 	switch msg {
// 	case "EVENTA":
// 		t.eventBus.Publish(eventbus.EventAType, eventbus.EventA.Add(map[string]any{"str": "ab1d", "int": 1, "arr": []int{1, 2, 3}}))
// 		logger.Println("[TCP] AAAAAAAAAAAAA")
// 	case "EVENTB":
// 		t.eventBus.Publish(eventbus.EventBType, eventbus.EventB.Add(eventbus.NewAddPayload().SetData(10)))
// 		logger.Println("[TCP] BBBBBBBBBBBBB")
// 	default:
// 		logger.Println("[TCP] Unknown Message Type")
// 	}
// }

func (t *TCPServer) Shutdown() {
	t.mu.Lock()
	t.cancel()
	// for len(t.msgChannel) > 0 {
	// 	<-t.msgChannel
	// }
	// close(t.msgChannel)
	if t.listener != nil {
		t.listener.Close()
	}

	// send 루틴 종료
	t.clients.Close()

	t.mu.Unlock()
	t.wg.Wait()
	logger.Println("[TCP] TCP goroutine terminated")
}

/**
 * If managing monitoring thread individually, use the function below, otherwise manage them in system monitoring.
 */

func (t *TCPServer) StartTCPServerHeartbeat() {
	t.wg.Add(1)
	go func() {
		heartbeat := time.NewTicker(t.heartbeat)
		healthCheck := time.NewTicker(t.heartbeat * 3)

		defer func() {
			heartbeat.Stop()
			healthCheck.Stop()
			t.wg.Done()
		}()

		for {
			select {
			case <-heartbeat.C:
				if !t.IsListening() {
					logger.Println("[TCP Server Heartbeat] Connection is closed, attempting to reconnect...")
					if err := t.Listening(); err != nil {
						logger.Printf("[TCP Server Heartbeat] Failed to reconnect:\n\t%v", err)
					}
				}
			case <-healthCheck.C:
				if t.IsListening() && !t.CheckConnection() {
					logger.Println("[TCP Server Heartbeat] Connection is closed, attempting to reconnect...")
					if err := t.Listening(); err != nil {
						logger.Printf("[TCP Server Heartbeat] Failed to reconnect:\n\t%v", err)
					}
				}
			case <-t.ctx.Done():
				logger.Println("[TCP Server Heartbeat] TCP Client heartbeat goroutine terminated")
				return
			}
		}
	}()
}

func (t *TCPServer) CheckConnection() bool {
	t.mu.RLock()
	listener := t.listener
	isListening := t.isListening
	t.mu.RUnlock()

	if listener == nil || !isListening {
		return false
	}

	testConn, err := net.DialTimeout("tcp", "localhost:5000", 300*time.Millisecond)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			if t.timeOutCnt >= t.maxTimeOutCnt {
				t.timeOutCnt = 0
				listener.Close()
				t.SetListeningState(false)
				return false
			}
			t.timeOutCnt++
			return false
		} else {
			t.SetListeningState(false)
			return false
		}
	}
	testConn.Close()

	return true
}

// for test
func (t *TCPServer) HeartbeatTest() {
	t.listener.Close()
	t.SetListeningState(false)
}
