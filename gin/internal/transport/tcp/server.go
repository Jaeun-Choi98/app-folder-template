package tcp

import (
	"bufio"
	"context"
	"net"
	"pjt/internal/logger"
	model "pjt/internal/model/event"
	"pjt/internal/transport/eventbus"
	"strings"
	"sync"
	"time"
)

/**
 * 서버에서 클라이언트를 별도로 관리하기 위한 구조체체
 */

type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient(conn net.Conn, parentCtx context.Context) *Client {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Client) ReadMessage() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

/**
 * TCP Server
 */

type TCPServer struct {
	listener    net.Listener
	Clients     map[*Client]bool
	EventBus    *eventbus.EventBus
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	wg          sync.WaitGroup
	reconnect   time.Duration
	isListening bool
}

func NewTCPServer(eventBus *eventbus.EventBus, reconnect time.Duration) (*TCPServer, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPServer{
		Clients:   make(map[*Client]bool),
		EventBus:  eventBus,
		reconnect: reconnect,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

func (t *TCPServer) Listening() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		logger.Println(err)
		return err
	}

	t.listener = listener
	t.isListening = true
	return nil
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
}

func (t *TCPServer) Start() error {
	// SendMessage goroutine을 닫기 위한 동기화
	// 클라이언트에게 다시 메시지를 보내야한다면 아래 코드 필요
	//wg.Add(1)
	//go SendMessage(ctx)

	// 초기 연결
	if err := t.Listening(); err != nil {
		return err
	}

	for {
		t.mu.RLock()
		listener := t.listener
		t.mu.RUnlock()

		if listener == nil {
			if t.IsListening() {
				t.SetListeningState(false)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-t.ctx.Done():
				return err
			default:
			}
			continue
		}
		t.wg.Add(1)
		go t.HandleConnection(t.ctx, conn)
	}
}

func (t *TCPServer) HandleConnection(ctx context.Context, conn net.Conn) {

	client := NewClient(conn, ctx)

	defer func() {
		conn.Close()
		client.cancel()
		t.unregisterClient(client)
		logger.Printf("[TCP Server] Disconnected client: %s\n", conn.RemoteAddr().String())
		t.wg.Done()
	}()

	t.registerClient(client)
	logger.Printf("[TCP Server] Connected new client: %s\n", conn.RemoteAddr().String())

	handleMessage := func(msg string) {
		switch msg {
		case "EVENTA":
			t.EventBus.Publish(model.EVENTA, &model.EventA{Type: model.EVENTA, Payload: model.MessageA{Name: "cju", Age: 11}})
			logger.Println("[TCP Server] AAAAAAAAAAAAA")
		case "EVENTB":
			t.EventBus.Publish(model.EVENTB, &model.EventB{Type: model.EVENTB, Payload: model.MessageB{Args: []string{"sdf", "sdf"}, Cmd: "stres"}})
			logger.Println("[TCP Server] BBBBBBBBBBBBB")
		default:
			logger.Println("[TCP Server] Unknown Message Type")
		}
	}

	receiveMessage := func() {
		receiveMsgChan := make(chan string, 1)
		defer func() {
			close(receiveMsgChan)
			//logger.Println("[TCP Server] Closed receiveMessage func")
		}()

		closedConnByRemoteHost := make(chan bool, 1)
		defer close(closedConnByRemoteHost)

		for {
			// if an existing connection was forcibly closed by the remote host, break select block
			go func() {
				msg, err := client.ReadMessage()
				if err != nil {
					logger.Println(err)
					closedConnByRemoteHost <- true
					return
				}
				receiveMsgChan <- msg
			}()

			select {
			case <-ctx.Done():
				return
			case <-closedConnByRemoteHost:
				return
			case msg := <-receiveMsgChan:
				handleMessage(msg)
			}
		}
	}

	receiveMessage()
}

func (t *TCPServer) registerClient(client *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Clients[client] = true
}

func (t *TCPServer) unregisterClient(client *Client) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.Clients, client)
}

func (t *TCPServer) Shutdown() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()
	t.wg.Wait()
	logger.Println("[TCP] TCP goroutine terminated")

	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

func (t *TCPServer) StartTCPServerHeartbeat() {
	go func() {
		heartbeat := time.NewTicker(t.reconnect)
		defer heartbeat.Stop()
		for {
			select {
			case <-heartbeat.C:
				if !t.IsListening() || !t.checkConnection() {
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

func (t *TCPServer) checkConnection() bool {
	t.mu.RLock()
	listener := t.listener
	isListening := t.isListening
	t.mu.RUnlock()

	if listener == nil || !isListening {
		return false
	}

	// 별도의 테스트 연결로 포트 상태 확인
	testConn, err := net.DialTimeout("tcp", "localhost:5000", 100*time.Millisecond)
	if err != nil {
		return false
	}
	testConn.Close()

	return true
}

// for test
func (t *TCPServer) HeartbeatTest() {
	t.listener.Close()
}

// func SendMessage(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			// 정상 종료 시에 SendMessage goroutine이 제대로 닫히는지 확인.
// 			log.Println("closed SendMessage goruntine")
// 			wg.Done()
// 			return
// 			// 여기에 이벤트 버스 구독 채널이 와야함
// 		case strMsg := <-messageChan:
// 			mu.RLock()
// 			for client, exists := range clients {
// 				if exists {
// 					client.conn.Write([]byte(strMsg))
// 				}
// 			}
// 			mu.RUnlock()
// 		}
// 	}
// }
