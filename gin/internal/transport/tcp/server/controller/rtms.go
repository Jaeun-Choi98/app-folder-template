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

func (c *TCPController) Handle0x001() handler.TypeHandlerFunc {
	return func(msg tcpmd.ParseMsg, replyCh map[any]chan tcpmd.Reply) error {
		rtmsMsg := msg.GetPacket().(*implparser.Packet)
		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return fmt.Errorf("invaild LRC")
		}
		log.Println("연결 성공")
		return nil
	}
}

// reply test
func (hm *TCPController) Handle0x002() handler.TypeHandlerFunc {
	return func(msg tcpmd.ParseMsg, replyCh map[any]chan tcpmd.Reply) error {
		rtmsMsg := msg.GetPacket().(*implparser.Packet)

		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return fmt.Errorf("invaild LRC")
		}
		replyCh[0x02] <- &tcpmd.GenericReply[string]{Err: nil, Payload: "success reply test"}
		return nil
	}
}

func (c *TCPController) Handle0x010() handler.TypeHandlerFunc {
	return func(msg tcpmd.ParseMsg, replyCh map[any]chan tcpmd.Reply) error {

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		c.TcpService.Handle0x010(msg.GetClientId())
		return nil
	}
}

// 클라이언트 종료 시에 TCP 세션 정보를 삭제하기 위한 핸들러
func (c *TCPController) Handle0xAA() handler.TypeHandlerFunc {
	return func(msg tcpmd.ParseMsg, replyCh map[any]chan tcpmd.Reply) error {
		c.TcpService.Handle0xAA(msg.GetClientId())
		return nil
	}
}
