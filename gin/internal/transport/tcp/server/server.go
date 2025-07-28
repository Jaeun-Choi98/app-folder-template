package tcp

import (
	"context"
	"encoding/binary"
	"net"
	"pjt/internal/logger"
	"pjt/internal/service"
	"pjt/internal/transport/tcp/server/client"
	"pjt/internal/transport/tcp/server/handler"
	"pjt/internal/transport/tcp/server/parser"
	"sync"
	"time"
)

type TCPServer struct {
	listener net.Listener
	clients  *client.ClientManager

	handler handler.HandlerManagerInterface

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	isListening bool
	heartbeat   time.Duration

	timeOutCnt    int
	maxTimeOutCnt int
}

func NewTCPServer(clientManager *client.ClientManager, tcpServcie service.TCPServiceInterface, heartbeat time.Duration) (*TCPServer, error) {

	ctx, cancel := context.WithCancel(context.Background())
	return &TCPServer{
		clients:       clientManager,
		handler:       handler.NewHandlerManager(tcpServcie),
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

	//c := client.NewClient(t.ctx, uuid.New().ID(), conn, parser.NewRTMSParser(binary.LittleEndian, 4096), t.handler)

	// sendWork Test
	c := client.NewClient(t.ctx, 1, conn, parser.NewRTMSParser(binary.LittleEndian, 4096), t.handler)

	defer func() {
		//logger.Println(c.ClientId)
		t.clients.Unregister(c.ClientId)
		c.Close()
		//logger.Printf("[TCP] Disconnected client: %s", c.Conn.RemoteAddr().String())
		t.wg.Done()
	}()

	t.clients.Register(c)
	//logger.Printf("[TCP] Connected new client: %s", c.Conn.RemoteAddr().String())
	c.MessageProcessingLoop()
}

func (t *TCPServer) Shutdown() {
	t.mu.Lock()
	t.cancel()
	if t.listener != nil {
		t.listener.Close()
	}

	// shutdown sendWorker routine
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
