package client

import (
	"context"
	"io"
	"net"
	"pjt/internal/logger"
	"sync"

	"pjt/internal/transport/tcp/server/handler"
	"pjt/internal/transport/tcp/server/parser"

	"time"
)

const (
	SequenceMode = 65535
)

/**
 * 서버에서 클라이언트를 별도로 관리하기 위한 구조체
 */

type Client struct {
	ClientId uint32
	Conn     net.Conn
	SeqNum   uint16

	parser  parser.Parser
	handler handler.HandlerManagerInterface

	ReplyCh map[byte]chan *handler.ReplyMessage // OPCODE -> Reply Channel
	//TimeoutChannel chan bool

	IsAuth bool

	parsingErrCnt int

	Ctx    context.Context
	Cancel context.CancelFunc

	mu sync.Mutex
}

func NewClient(parentCtx context.Context, clinetId uint32, conn net.Conn, ps parser.Parser,
	hd handler.HandlerManagerInterface) *Client {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Client{
		ClientId: clinetId,
		Conn:     conn,
		parser:   ps,
		handler:  hd,
		ReplyCh:  make(map[byte]chan *handler.ReplyMessage),
		//TimeoutChannel: make(chan bool, 1),
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (c *Client) Close() {
	// if c.IsAuth {
	// 	logger.Printf("[TCP] Disconnected client: %d", c.ClientId)
	// }
	if c.ClientId < 65536 {
		logger.Printf("[TCP] Disconnected client: %d", c.ClientId)
	}
	if c.Conn != nil {
		c.Conn.Close()
	}
	for _, v := range c.ReplyCh {
		// 클라이언트가 패닉이 발생했을 때, 해당 replyCh를 갑자기 닫으면 해당 채널을 수신하고 있는 곳에서 close of closed channel panic 발생. -> 수신하는 곳에서 recover 처리 필요.
		// client_manager.go_191
		close(v)
	}
	//close(c.TimeoutChannel)
	c.Cancel()
}

func (c *Client) MessageProcessingLoop() {
	for {
		select {
		case <-c.Ctx.Done():
			return
		default:
			msg, err := c.parser.Parse(c.Conn)

			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				// if EOF, normally exit
				if err == io.EOF || c.parsingErrCnt > 4 {
					return
				}
				logger.Printf("[TCP] Parse error for client %d: %v", c.ClientId, err)
				c.parsingErrCnt++
				continue
			}

			msg.ClientId = c.ClientId

			if err := c.handler.HandleMessage(msg, c.ReplyCh); err != nil {
				logger.Printf("[TCP] Handle error: %v", err)
				// 클라이언트에게 처리 결과를 반환해야 한다면 아래 코드 실행.
				// c.SendMessage([]byte(err.Error()), 0)
			}
			// c.SendMessage([]byte("success"), 0)
		}
	}
}

func (c *Client) SendMessage(msg []byte, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	c.Conn.SetWriteDeadline(time.Now().Add(timeout))
	defer c.Conn.SetWriteDeadline(time.Time{})

	written := 0
	for written < len(msg) {
		n, err := c.Conn.Write(msg[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

func (c *Client) ReadMessage() error {
	return nil
}

func (c *Client) SetSequenceNum(seq uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SeqNum = seq
}

func (c *Client) GetSequenceNum() uint16 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.SeqNum
}
