package client

import (
	"context"
	"io"
	"net"
	"pjt/internal/logger"
	"sync"

	implparser "pjt/internal/transport/tcp/server/parser"

	"github.com/Jaeun-Choi98/modules/tcpnet/server/handler"
	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/server/model"
	"github.com/Jaeun-Choi98/modules/tcpnet/server/parser"

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

	handler handler.ManagerInterface

	//ReplyCh map[any]chan tcpmd.Reply // OPCODE -> Reply Channel
	//TimeoutChannel chan bool

	IsAuth bool

	parsingErrCnt int

	// 아래 필드는 패키지를 사용한 예시
	parser parser.Parser

	wg sync.WaitGroup

	Ctx           context.Context
	Cancel        context.CancelFunc
	clientContext *tcpmd.ClientContext
	mu            sync.Mutex
}

func NewClient(parentCtx context.Context, clinetId uint32, conn net.Conn, ps *implparser.RTMSParser,
	hd handler.ManagerInterface) *Client {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Client{
		ClientId: clinetId,
		Conn:     conn,
		//parser:   ps,
		parser:  ps,
		handler: hd,
		//ReplyCh: make(map[any]chan tcpmd.Reply),
		clientContext: tcpmd.NewClientContext(context.Background(), 30),
		//TimeoutChannel: make(chan bool, 1),
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (c *Client) Close() {
	c.Cancel()
	c.wg.Wait()
	if c.IsAuth {
		logger.Printf("[TCP] Disconnected client: %d", c.ClientId)
		c.clientContext.SetParsedMsg(&implparser.ParseMessage{Type: "rtms_0x0AA", ClientId: c.ClientId})
		c.handler.HandleMessageSync(c.clientContext)
	}

	if c.Conn != nil {
		c.Conn.Close()
	}

	c.clientContext.Close()
}

func (c *Client) MessageProcessingLoop() {
	c.wg.Add(1)
	defer c.wg.Done()
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

			msg.SetClientId(c.ClientId)

			c.clientContext.SetParsedMsg(msg)

			if err := c.handler.HandleMessage(c.clientContext); err != nil {
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
