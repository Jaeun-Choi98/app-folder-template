package handler

import (
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/tcp/server/parser"
)

func HandleRTMSTrainWarning(evnetBus *eventbus.EventBus) TypeHandlerFunc {
	return func(msg *parser.BaseMessage) error {

		// warningData := map[string]interface{}{
		// 	"sequence":  rtmsMsg.Sequence,
		// 	"unit_no":   rtmsMsg.UnitNo,
		// 	"op_code":   rtmsMsg.OpCode,
		// 	"data":      rtmsMsg.Data2,
		// 	"client_id": msg.GetClientID(),
		// }

		// EventBus로 전파
		evnetBus.Publish(eventbus.EventAType, eventbus.NewEvent("EVENTA").Add(map[string]interface{}{"test": "sdfsd"}))

		// tcpService로 처리
		return nil
	}

}
