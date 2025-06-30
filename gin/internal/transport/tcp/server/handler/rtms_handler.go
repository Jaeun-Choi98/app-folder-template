package handler

import (
	"log"
	"pjt/internal/transport/tcp/server/parser"
)

func (hm *HandlerManager) Handle0x010() TypeHandlerFunc {
	return func(msg *parser.BaseMessage) error {
		rtmsMsg, ok := msg.Data.(*parser.RTMSMessage)
		if !ok {
			return errInvalidAssertion
		}

		var valid uint8
		valid += uint8(rtmsMsg.Length)
		valid += rtmsMsg.Sequence
		valid += rtmsMsg.UnitNo
		valid += uint8(rtmsMsg.OpCode)

		var id, passwd string
		for i := 0; i < 40; i++ {
			valid += rtmsMsg.Data[i]
			if i < 20 {
				id += string(rtmsMsg.Data[i])
			} else {
				passwd += string(rtmsMsg.Data[i])
			}
		}

		valid = (valid ^ 0xFF) | 0x01

		log.Println(id, passwd, valid, rtmsMsg.OpCode, rtmsMsg.Length, valid)

		// LRC 검증 이렇게 하는게 맞는지 모름.
		// if valid != rtmsMsg.LRC {
		// 	return errInvalidLRC
		// }

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		hm.TCPService.Handle0x010()
		return nil
	}
}
