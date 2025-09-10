package tcp

import (
	"context"
	"encoding/binary"
	"net"
	"pjt/internal/config"
	"pjt/internal/logger"
	"pjt/internal/transport/tcp/server/client"
	tcpcontroller "pjt/internal/transport/tcp/server/controller"

	implparser "pjt/internal/transport/tcp/server/parser"
	"time"

	mhandler "github.com/Jaeun-Choi98/modules/tcpnet/server/handler"
	"github.com/Jaeun-Choi98/modules/tcpnet/server/server"
)

/**
 * tcpnet.server 패키지를 사용한 구현 예시 템플릿
 */

type TCPServer struct {
	*server.ServerBase
	clients *client.ClientManager
	router  mhandler.HandlerManagerInterface
}

func NewTCPServer(tcpcontroller *tcpcontroller.TCPController, cm *client.ClientManager, cfg *config.Configuration, heartbeat time.Duration, maxTimeOutCnt int) (*TCPServer, error) {
	s, err := server.NewServerBase(context.Background(), heartbeat, maxTimeOutCnt)
	if err != nil {
		return nil, err
	}
	s.SetIpAndPort(cfg.TcpIp, cfg.TcpPort)
	customServer := &TCPServer{
		ServerBase: s,
		clients:    cm,
		router:     tcpcontroller.Router,
	}
	customServer.ImplClientHandleConnection()
	return customServer, nil
}

func (s *TCPServer) ImplClientHandleConnection() {
	s.ServerBase.SetHandleConnectFunc(func(conn net.Conn) {
		//c := client.NewClient(t.ctx, uuid.New().ID(), conn, parser.NewRTMSParser(binary.LittleEndian, 4096), t.handler)

		// sendWork Test
		c := client.NewClient(context.Background(), 1, conn, implparser.NewRTMSParser(binary.LittleEndian, 4096), s.router)

		defer func() {
			//logger.Println(c.ClientId)
			s.clients.Unregister(c.ClientId)
			c.Close()
			//logger.Printf("[TCP] Disconnected client: %s", c.Conn.RemoteAddr().String())

		}()

		s.clients.Register(c)
		//logger.Printf("[TCP] Connected new client: %s", c.Conn.RemoteAddr().String())
		c.MessageProcessingLoop()
	})
}

func (s *TCPServer) Start() error {
	return s.ServerBase.Start()
}

func (s *TCPServer) Shutdown() {
	s.ServerBase.Shutdown()
	logger.Println("[TCP] TCP goroutine terminated")
}
