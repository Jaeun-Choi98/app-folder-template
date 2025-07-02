package handler

import (
	"log"
	"pjt/internal/transport/tcp/server/parser"
	"pjt/internal/transport/tcp/server/serializer"
)

func (hm *HandlerManager) Handle0x010() TypeHandlerFunc {
	return func(msg *parser.BaseMessage) error {
		rtmsMsg, ok := msg.Data.(*parser.RTMSMessage)
		if !ok {
			return errInvalidAssertion
		}
		var id, passwd string
		for i := 0; i < 40; i++ {
			if i < 20 {
				id += string(rtmsMsg.Data[i])
			} else {
				passwd += string(rtmsMsg.Data[i])
			}
		}

		lrcVerifiy := serializer.CalculateRTMSLRC(rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)
		log.Println("sdf")
		if lrcVerifiy != rtmsMsg.LRC {
			return errInvalidLRC
		}

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		hm.TCPService.Handle0x010()
		return nil

	}
}
