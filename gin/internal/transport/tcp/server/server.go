package tcp

import (
	"bufio"
	"context"
	"log"
	"net"
	"pjt/internal/logger"
	"pjt/internal/transport/eventbus"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

/**
 * 서버에서 클라이언트를 별도로 관리하기 위한 구조체체
 */

type TCPClient struct {
	clientId string
	conn     net.Conn
	reader   *bufio.Reader

	msgChannel     chan string
	timeoutChannel chan bool

	ctx    context.Context
	cancel context.CancelFunc
}

func NewTCPClient(parentCtx context.Context, clinetId string, conn net.Conn) *TCPClient {
	ctx, cancel := context.WithCancel(parentCtx)
	return &TCPClient{
		clientId:       clinetId,
		conn:           conn,
		reader:         bufio.NewReader(conn),
		msgChannel:     make(chan string, 10),
		timeoutChannel: make(chan bool, 1),
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (c *TCPClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.msgChannel)
	close(c.timeoutChannel)
	c.reader = nil
	c.cancel()
}

func (c *TCPClient) ReadMessage() (string, error) {
	line, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

type TCPServer struct {
	listener   net.Listener
	clients    map[string]*TCPClient
	msgChannel chan string

	eventBus *eventbus.EventBus

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	isListening bool
	heartbeat   time.Duration
}

func NewTCPServer(eventBus *eventbus.EventBus, heartbeat time.Duration) (*TCPServer, error) {

	ctx, cancel := context.WithCancel(context.Background())

	return &TCPServer{
		clients:    make(map[string]*TCPClient),
		msgChannel: make(chan string, 10),
		eventBus:   eventBus,
		ctx:        ctx,
		cancel:     cancel,
		heartbeat:  heartbeat,
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

func (t *TCPServer) registerClient(clientId string, client *TCPClient) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.clients[clientId]; !exists {
		t.clients[clientId] = client
	}
}

func (t *TCPServer) unregisterClient(clientId string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.clients, clientId)
}

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

	client := NewTCPClient(t.ctx, uuid.New().String(), conn)
	defer func() {
		t.unregisterClient(client.clientId)
		client.Close()
		logger.Printf("[TCP] Disconnected client: %s", client.conn.RemoteAddr().String())
		t.wg.Done()
	}()

	t.registerClient(client.clientId, client)
	logger.Printf("[TCP] Connected new client: %s", client.conn.RemoteAddr().String())
	for {
		go func() {
			if tcpClient, ok := client.conn.(*net.TCPConn); ok {
				tcpClient.SetDeadline(time.Now().Add(30 * time.Second))
			}
			msg, err := client.ReadMessage()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					client.timeoutChannel <- true
					return
				}
				logger.Println(err)
				client.cancel()
				return
			}
			client.msgChannel <- msg
		}()

		select {
		case <-ctx.Done():
			return
		case <-client.ctx.Done():
			return
		case <-client.timeoutChannel:
			logger.Printf("[TCP] Clinet conn is timeout: %s", client.conn.RemoteAddr().String())
		case msg := <-client.msgChannel:
			t.handleMessage(msg)
		}
	}
}

// if need to send to particular client, add parameter clientId
func (t *TCPServer) handleMessage(msg string) {
	switch msg {
	case "EVENTA":
		t.eventBus.Publish(eventbus.EVENTA, &eventbus.EventA{Type: "EVENTA", Payload: eventbus.PayloadA{Name: "cju", Age: 11}})
		logger.Println("[TCP] AAAAAAAAAAAAA")
	case "EVENTB":
		t.eventBus.Publish(eventbus.EVENTB, &eventbus.EventB{Type: "EVENTB", Payload: eventbus.PayloadB{Args: []string{"sdf", "sdf"}, Cmd: "stres"}})
		logger.Println("[TCP] BBBBBBBBBBBBB")
	default:
		logger.Println("[TCP] Unknown Message Type")
	}
}

// broadcast
func (t *TCPServer) SendMessageToAllRoutine() {
	for {
		select {
		case <-t.ctx.Done():
			// 정상 종료 시에 SendMessage goroutine이 제대로 닫히는지 확인.
			log.Println("[TCP] SendMessage goroutine is terminated")
			t.wg.Done()
			return
		case strMsg := <-t.msgChannel:
			t.mu.RLock()
			for _, client := range t.clients {
				client.conn.Write([]byte(strMsg))
			}
			t.mu.RUnlock()
		}
	}
}

func (t *TCPServer) SendMessageToAllOnce(msg string) {
	t.mu.RLock()
	for _, client := range t.clients {
		client.conn.Write([]byte(msg))
	}
	t.mu.RUnlock()
}

func (t *TCPServer) SendMessageToClient(msg string, clinetId string) {
	t.mu.RLock()
	if client, exists := t.clients[clinetId]; exists {
		client.conn.Write([]byte(msg))
	}
	t.mu.RUnlock()
}

func (t *TCPServer) Shutdown() {
	t.mu.Lock()
	t.cancel()
	for len(t.msgChannel) > 0 {
		<-t.msgChannel
	}
	close(t.msgChannel)
	if t.listener != nil {
		t.listener.Close()
	}
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

		defer func() {
			heartbeat.Stop()
			t.wg.Done()
		}()

		for {
			select {
			case <-heartbeat.C:
				if !t.IsListening() || !t.CheckConnection() {
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

	testConn, err := net.DialTimeout("tcp", "localhost:5000", 100*time.Millisecond)
	if err != nil {
		t.SetListeningState(false)
		return false
	}
	testConn.Close()

	return true
}

// for test
func (t *TCPServer) HeartbeatTest() {
	t.listener.Close()
	t.SetListeningState(false)
}
