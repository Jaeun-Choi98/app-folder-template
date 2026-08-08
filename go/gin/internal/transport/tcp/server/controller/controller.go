package tcpcontroller

import (
	"encoding/binary"
	"fmt"
	"log"
	"pjt/internal/service"
	implparser "pjt/internal/transport/tcp/server/parser"
	"pjt/internal/transport/tcp/server/serializer"

	"github.com/Jaeun-Choi98/modules/tcpnet/advanced/handler"
	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/advanced/model"
)

func NewController(svc service.TCPServiceInterface) handler.ManagerInterface {
	h := handler.New[string]()

	h.RegisterHandle("text", func(c *tcpmd.HandleContext) error {
		log.Printf("msgType: %s, clientID: %d, Data: %v ", c.GetParseMsg().GetPacketId(),
			c.GetParseMsg().GetClientId(), c.GetParseMsg().GetPacket())
		return nil
	})

	h.RegisterHandle("json", func(c *tcpmd.HandleContext) error {
		return nil
	})

	h.RegisterHandle("rtms_0x010", handle0x010(svc))
	// 연결 상태 확인
	h.RegisterHandle("rmts_0x001", handle0x001())
	// reply test
	h.RegisterHandle("rtms_0x002", handle0x002())
	// close client
	h.RegisterHandle("rtms_0x0AA", handle0xAA(svc))

	return h
}

func handle0x001() handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {

		rtmsMsg := c.GetParseMsg().GetPacket().(*implparser.Packet)
		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return fmt.Errorf("invaild LRC")
		}
		log.Println("연결 성공")
		return nil
	}
}

// reply test
func handle0x002() handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {
		rtmsMsg := c.GetParseMsg().GetPacket().(*implparser.Packet)

		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return fmt.Errorf("invaild LRC")
		}
		ch, ok := c.GetReplyChannel().Get(0x02)
		if ok {
			ch <- &tcpmd.GenericReply[string]{Err: nil, Payload: "success reply test"}
		}
		return nil
	}
}

func handle0x010(svc service.TCPServiceInterface) handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		svc.Handle0x010(c.GetParseMsg().GetClientId())
		return nil
	}
}

// 클라이언트 종료 시에 TCP 세션 정보를 삭제하기 위한 핸들러
func handle0xAA(svc service.TCPServiceInterface) handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {
		svc.Handle0xAA(c.GetParseMsg().GetClientId())
		return nil
	}
}
