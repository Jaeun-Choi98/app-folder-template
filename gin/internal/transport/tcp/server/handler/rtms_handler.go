package handler

import (
	"encoding/binary"
	"log"
	"pjt/internal/transport/tcp/server/parser"
	"pjt/internal/transport/tcp/server/serializer"
)

func (hm *HandlerManager) Handle0x001() TypeHandlerFunc {
	return func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		rtmsMsg := msg.Packet
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
		rtmsMsg := msg.Packet

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

		/**
		 * EventBus로 전파하거나  tcpService로 처리
		 */
		hm.TCPService.Handle0x010(msg.ClientId)
		return nil
	}
}

// 클라이언트 종료 시에 TCP 세션 정보를 삭제하기 위한 핸들러
func (hm *HandlerManager) Handle0xAA() TypeHandlerFunc {
	return func(msg *parser.ParseMessage, replyCh map[byte]chan *ReplyMessage) error {
		hm.TCPService.Handle0xAA(msg.ClientId)
		return nil
	}
}
