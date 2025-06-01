package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// Configuration 객체 주입 필요 ( ip, port )
type TCPClient struct {
	conn        net.Conn
	sendMsgChan chan string
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	isConnected bool
	timeout     time.Duration
	reconnect   time.Duration
}

func NewTCPClient(timeout, reconnect time.Duration) (*TCPClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPClient{
		timeout:     timeout,
		reconnect:   reconnect,
		sendMsgChan: make(chan string),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (t *TCPClient) Connect() error {

	t.mu.Lock()
	defer t.mu.Unlock()

	conn, err := net.DialTimeout("tcp", "localhost:5000", t.timeout)
	if err != nil {
		log.Println(err)
		return err
	}

	t.conn = conn
	t.isConnected = true
	return nil
}

func (t *TCPClient) Start() error {

	// 초기 연결
	if err := t.Connect(); err != nil {
		return err
	}

	t.wg.Add(1)
	defer t.wg.Done()
	// 이후에 receive 함수로 분리 필요.
	reader := bufio.NewReader(os.Stdin)
	for {
		go func() {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}
			t.sendMsgChan <- line
		}()

		select {
		case <-t.ctx.Done():
			return nil
		case data := <-t.sendMsgChan:
			err := t.SendMessage(data)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
}

func (t *TCPClient) SendMessage(msg string) error {

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isConnected {
		return fmt.Errorf("not connected")
	}

	byteMsg := append([]byte(msg), '\n')
	_, err := t.conn.Write(byteMsg)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (t *TCPClient) Shutdown() error {
	t.cancel()
	t.wg.Wait()
	log.Println("[TCP Client] TCP Client goroutine terminated")
	t.isConnected = false
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

func (t *TCPClient) StartTCPClientHeartbeat() {
	go func() {
		heartbeat := time.NewTicker(t.reconnect)
		defer heartbeat.Stop()
		for {
			select {
			case <-heartbeat.C:
				connected := t.isConnected && t.checkConnection()
				if !connected {
					log.Println("[TCP Client Heartbeat] Connection is closed, attempting to reconnect...")
					if err := t.Connect(); err != nil {
						log.Printf("[TCP Client Heartbeat] Failed to reconnect:\n\t%v", err)
					}
				}
			case <-t.ctx.Done():
				log.Println("[TCP Client Heartbeat] TCP Client heartbeat goroutine terminated")
				return
			}
		}
	}()
}

// check the connection
func (t *TCPClient) checkConnection() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.conn == nil {
		return false
	}

	// TCP 연결 상태 확인을 위한 더미 데이터 전송
	t.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	_, err := t.conn.Write([]byte{})
	if err != nil {
		t.isConnected = false
		return false
	}

	return true
}

// client test
// func main() {
// 	tcpClient, err := NewTCPClient()
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	signalChan := make(chan os.Signal, 1)
// 	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
// 	go tcpClient.Start()

// 	<-signalChan
// 	log.Println("Shutting down...")
// 	tcpClient.Shutdown()
// 	log.Println("TCP Client terminated")
// 	close(signalChan)
// }
