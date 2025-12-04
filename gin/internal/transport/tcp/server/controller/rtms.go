package tcpcontroller

import (
	"encoding/binary"
	"fmt"
	"log"
	implparser "pjt/internal/transport/tcp/server/parser"
	"pjt/internal/transport/tcp/server/serializer"

	"github.com/Jaeun-Choi98/modules/tcpnet/server/handler"
	tcpmd "github.com/Jaeun-Choi98/modules/tcpnet/server/model"
)

func (ctr *TCPController) Handle0x001() handler.HandlerFunc {
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
func (ctr *TCPController) Handle0x002() handler.HandlerFunc {
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

func (ctr *TCPController) Handle0x010() handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		ctr.TcpService.Handle0x010(c.GetParseMsg().GetClientId())
		return nil
	}
}

// 클라이언트 종료 시에 TCP 세션 정보를 삭제하기 위한 핸들러
func (ctr *TCPController) Handle0xAA() handler.HandlerFunc {
	return func(c *tcpmd.HandleContext) error {
		ctr.TcpService.Handle0xAA(c.GetParseMsg().GetClientId())
		return nil
	}
}
