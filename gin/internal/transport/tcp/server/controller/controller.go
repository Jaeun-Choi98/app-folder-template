package tcpcontroller

import (
	"log"
	"pjt/internal/service"

	"github.com/Jaeun-Choi98/modules/tcpnet/server/handler"
	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/server/model"
)

type TCPController struct {
	Router     *handler.Manager[string]
	TcpService service.TCPServiceInterface
}

func NewTCPController(router *handler.Manager[string], tcpService service.TCPServiceInterface) *TCPController {
	controller := &TCPController{
		Router:     router,
		TcpService: tcpService,
	}
	controller.RoutePath()
	return controller
}

func (c *TCPController) RoutePath() {
	c.Router.RegisterHandle("text", func(c *tcpmd.ReqContext) error {
		log.Printf("msgType: %s, clientID: %d, Data: %v ", c.GetParsedMsg().GetPacketId(),
			c.GetParsedMsg().GetClientId(), c.GetParsedMsg().GetPacket())
		return nil
	})

	c.Router.RegisterHandle("json", func(c *tcpmd.ReqContext) error {
		return nil
	})

	c.Router.RegisterHandle("rtms_0x010", c.Handle0x010())
	// 연결 상태 확인
	c.Router.RegisterHandle("rmts_0x001", c.Handle0x001())
	// reply test
	c.Router.RegisterHandle("rtms_0x002", c.Handle0x002())
	// close client
	c.Router.RegisterHandle("rtms_0x0AA", c.Handle0xAA())
}
