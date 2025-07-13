package handler

import (
	"encoding/binary"
	"log"
	"pjt/internal/transport/tcp/server/parser"
	"pjt/internal/transport/tcp/server/serializer"
)

func (hm *HandlerManager) Handle0x001() TypeHandlerFunc {
	return func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		rtmsMsg, ok := msg.Data.(*parser.RTMSMessage)
		if !ok {
			return errInvalidAssertion
		}
		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return errInvalidLRC
		}
		log.Println("연결 성공")
		return nil
	}
}

// reply test
func (hm *HandlerManager) Handle0x002() TypeHandlerFunc {
	return func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		rtmsMsg, ok := msg.Data.(*parser.RTMSMessage)
		if !ok {
			return errInvalidAssertion
		}
		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)

		if lrcVerifiy != rtmsMsg.LRC {
			return errInvalidLRC
		}
		replyCh[0x02] <- &ReplyMessage{Err: nil, Payload: "success reply test"}
		return nil
	}
}

func (hm *HandlerManager) Handle0x010() TypeHandlerFunc {
	return func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
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

		lrcVerifiy := serializer.CalculateRTMSLRC(binary.LittleEndian, rtmsMsg.Length, rtmsMsg.Sequence, rtmsMsg.UnitNo, rtmsMsg.OpCode, rtmsMsg.Data)
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
