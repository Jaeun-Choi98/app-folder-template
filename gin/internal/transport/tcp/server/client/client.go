package client

import (
	"context"
	"io"
	"net"
	"pjt/internal/logger"
	"pjt/internal/transport/tcp/server/handler"
	"pjt/internal/transport/tcp/server/parser"
)

/**
 * 서버에서 클라이언트를 별도로 관리하기 위한 구조체체
 */

type Client struct {
	ClientId string
	Conn     net.Conn

	parser  parser.Parser
	handler handler.HandlerManagerInterface

	//TimeoutChannel chan bool

	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewClient(parentCtx context.Context, clinetId string, conn net.Conn, ps parser.Parser,
	hd handler.HandlerManagerInterface) *Client {
	ctx, cancel := context.WithCancel(parentCtx)
	return &Client{
		ClientId: clinetId,
		Conn:     conn,
		parser:   ps,
		handler:  hd,
		//TimeoutChannel: make(chan bool, 1),
		Ctx:    ctx,
		Cancel: cancel,
	}
}

func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Close()
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
				// if EOF, normally exit
				if err == io.EOF {
					return
				}
				logger.Printf("[TCP] Parse error for client %s: %v", c.ClientId, err)
				continue
			}

			msg.ClientId = c.ClientId

			if err := c.handler.HandleMessage(msg); err != nil {
				logger.Printf("[TCP] Handle error: %v", err)
				c.SendMessage([]byte(err.Error()))
			}
		}
	}
}

func (c *Client) SendMessage(msg []byte) error {
	// 타임아웃 설정
	//c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := c.Conn.Write(msg)
	return err
}

func (c *Client) ReadMessage() error {
	return nil
}
