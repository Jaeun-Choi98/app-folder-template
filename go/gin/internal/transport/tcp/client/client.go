package tcp

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Jaeun-Choi98/modules/tcpnet/basic/client"
)

/**
 * client 패키지를 사용한 예시 구현 템플릿
 */
type CustomClient struct {
	BaseClient   *client.ClientBase
	firstRequest bool

	parsingErrCnt int

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewCustomClient(parentCtx context.Context, timeout, reconnect time.Duration) *CustomClient {
	c, _ := client.NewClientBase(parentCtx, timeout, reconnect)
	ctx, cancel := context.WithCancel(parentCtx)
	c.SetIpAndPort("localhost", "5000")
	customClient := &CustomClient{
		BaseClient:   c,
		firstRequest: true,
		ctx:          ctx,
		cancel:       cancel,
	}
	customClient.implHandleConnection()
	return customClient
}

func (c *CustomClient) implHandleConnection() {
	c.BaseClient.SetHandleConnectFunc(func(conn net.Conn) {
		defer c.wg.Done()
		for {

			select {
			case <-c.ctx.Done():
				return
			default:
			}

			if !c.BaseClient.IsConnected() {
				time.Sleep(1 * time.Second)
				continue
			}

			conn.SetReadDeadline(time.Now().Add(5 * time.Second))

			// =============== parser space =============== //
			parsedPacket, err := bufio.NewReader(conn).ReadString('\n')
			// =============== parser space =============== //

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// if EOF, normally exit
				if err == io.EOF || c.parsingErrCnt > 4 {
					c.BaseClient.SetConnectionState(false)
					return
				}
				c.parsingErrCnt++
				continue
			}

			// =============== handler space =============== //
			log.Print(parsedPacket)
			// =============== handler space =============== //
		}
	})
}

func (c *CustomClient) Start() error {
	c.wg.Add(1)
	go c.BaseClient.Start()
	return nil
}

func (c *CustomClient) Shutdown() {
	c.cancel()
	c.BaseClient.Shutdown()
	c.wg.Wait()
}

func (c *CustomClient) StartTCPClientHeartbeat() {
	c.BaseClient.StartTCPClientHeartbeat()
}

/**
 * 아래는 구현된 client 패키지 이전에 사용된 코드. 현재 사용 및 수정 (x)
 * deprecated
 */
type TCPClient struct {
	conn           net.Conn
	reader         *bufio.Reader
	receiveMsgChan chan string
	timeoutChan    chan bool

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex

	isConnected bool
	timeout     time.Duration
	reconnect   time.Duration
}

func NewTCPClient(timeout, reconnect time.Duration) (*TCPClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &TCPClient{
		//reader:      bufio.NewReader(os.Stdin),
		timeout:        timeout,
		reconnect:      reconnect,
		receiveMsgChan: make(chan string, 10),
		timeoutChan:    make(chan bool, 1),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func (c *TCPClient) SetConnectionState(state bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.isConnected = state
}

func (c *TCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.Unlock()
	return c.isConnected
}

func (c *TCPClient) Connect() error {

	c.mu.Lock()
	defer c.mu.Unlock()

	conn, err := net.DialTimeout("tcp", "localhost:5000", c.timeout)
	if err != nil {
		log.Println(err)
		return err
	}

	c.conn = conn
	c.isConnected = true
	c.reader = bufio.NewReader(c.conn)
	return nil
}

func (c *TCPClient) Start() error {

	// 초기 연결
	if err := c.Connect(); err != nil {
		return err
	}

	c.wg.Add(1)
	return c.WaitForReceiveMessage()
}

func (c *TCPClient) WaitForReceiveMessage() error {
	defer c.wg.Done()
	for {

		select {
		case <-c.ctx.Done():
			return nil
		default:
		}

		if !c.IsConnected() {
			time.Sleep(1 * time.Second)
			continue
		}

		go c.ReceiveMessage()

		select {
		case <-c.ctx.Done():
			return nil
		case data := <-c.receiveMsgChan:
			c.HandleMessage(data)
		case <-c.timeoutChan:
			log.Println("[TCP Client] Receive connection is timeout")
		}
	}
}

func (c *TCPClient) ReceiveMessage() {

	if tcpConn, ok := c.conn.(*net.TCPConn); ok {
		tcpConn.SetDeadline(time.Now().Add(30 * time.Second))
	}

	line, err := c.reader.ReadString('\n')
	if err != nil {
		if tcpErr, ok := err.(net.Error); ok && tcpErr.Timeout() {
			c.timeoutChan <- true
			return
		}
		log.Println(err)
		c.SetConnectionState(false)
		return
	}

	c.receiveMsgChan <- line
}

// HandleMessage is processing the received message.
func (t *TCPClient) HandleMessage(msg string) {
	t.SendMessage(msg)
}

func (t *TCPClient) SendMessage(msg string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	byteMsg := append([]byte(msg), '\n')
	_, err := t.conn.Write(byteMsg)
	if err != nil {
		log.Println(err)
		t.SetConnectionState(false)
		return err
	}
	return nil
}

func (t *TCPClient) Shutdown() error {
	t.cancel()

	for len(t.receiveMsgChan) > 0 {
		<-t.receiveMsgChan
	}
	close(t.receiveMsgChan)

	for len(t.timeoutChan) > 0 {
		<-t.timeoutChan
	}
	close(t.timeoutChan)

	t.isConnected = false
	if t.conn != nil {
		return t.conn.Close()
	}

	t.wg.Wait()
	log.Println("[TCP Client] TCP Client goroutine is terminated")
	return nil
}

func (t *TCPClient) StartTCPClientHeartbeat() {
	t.wg.Add(1)
	go func() {
		heartbeat := time.NewTicker(t.reconnect)

		defer func() {
			heartbeat.Stop()
			t.wg.Done()
		}()

		for {
			select {
			case <-heartbeat.C:
				connected := t.IsConnected() && t.CheckConnection()
				if !connected {
					log.Println("[TCP Client Heartbeat] Connection is closed, attempting to reconnect...")
					if err := t.Connect(); err != nil {
						log.Printf("[TCP Client Heartbeat] Failed to reconnect:\n\t%v", err)
					}
				}
			case <-t.ctx.Done():
				log.Println("[TCP Client Heartbeat] TCP Client heartbeat goroutine is terminated")
				return
			}
		}
	}()
}

// check the connection
func (t *TCPClient) CheckConnection() bool {
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
