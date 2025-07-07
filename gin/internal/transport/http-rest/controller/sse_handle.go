package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"pjt/internal/logger"
	"pjt/internal/service/sse-service/sse"
	"pjt/internal/transport/eventbus"
	"pjt/internal/transport/http-rest/http-utils/httperr"
	"pjt/internal/transport/http-rest/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/**
 * 클라이언트의 SSE 연결 요청을 처리 하기 위한 핸들러
 * 사용자의 ID가 없다면, ClientId를 UserId로 사용.
 *
 * 사용자ID가 없다면:
 * -> 특정 사용자에게만 데이터를 전송하는 /login/sse-send/:id API를 사용할 수 없음.
 * -> 하나의 세션에 하나의 클라이언트만 매칭 됨.
 */
func (ctl *Controller) HandleSSEConnect(ctx *gin.Context) {

	clientId := uuid.New().String()
	userId := ctx.GetString("id")
	if userId == "" {
		userId = clientId
	}

	client, err := ctl.SseService.NewSSEClient(clientId, userId, ctx)
	if err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	session := ctl.SseService.GetSessionByUserId(userId)
	if session == nil {
		session = ctl.SseService.NewSession(userId)
	}

	// 세션에 클라이언트 추가
	session.AddClient(clientId, client)

	// 연결 성공 이벤트 전송
	connectMsg := sse.Message{
		Type:    "connect",
		Payload: map[string]string{"message": "Connected successfully", "client_id": clientId},
	}
	client.SendMessage(connectMsg)

	eventChA := ctl.EventBus.Subscribe(eventbus.EventAType)
	//eventChB := ctl.EventBus.Subscribe(modelevent.EVENTB)
	eventSysSttCh := ctl.EventBus.Subscribe(eventbus.SysSttType)

	// 클라이언트 요청이 종료되면 정리
	// ctx.Request.Context()는 클라이언트가 연결을 끊으면 취소됩니다
	go func() {
		<-client.Ctx.Done()
		session.RemoveClient(clientId)
		ctl.EventBus.Unsubscribe(eventbus.EventAType, eventChA)
		if session.Count() == 0 {
			ctl.SseService.RemoveSession(userId)
		}
	}()

	sendMessage := func(event *eventbus.Event) error {
		payloadByte, err := json.Marshal(event.Payload)
		var msg sse.Message
		if err != nil {
			logger.Println("[SSE] failed to serialize event payload to json")
			msg = sse.Message{
				Type:    event.Type,
				Payload: event.Payload,
			}
		} else {
			msg = sse.Message{
				Type:    event.Type,
				Payload: string(payloadByte),
			}
		}
		if err := client.SendMessage(msg); err != nil {
			// 오류 발생 시 연결 종료
			session.RemoveClient(clientId)
			return err
		}
		return nil
	}

	// 클라이언트가 연결을 끊을 때까지 연결 유지
	for {
		select {
		case <-client.Ctx.Done():
			// 클라이언트 또는 서버에서 연결을 닫았음
			return
		case event := <-eventChA:
			if err := sendMessage(event); err != nil {
				return
			}
		case event := <-eventSysSttCh:
			if err := sendMessage(event); err != nil {
				return
			}
		}
	}
}

// SendEventToUser는 모든 사용자에게 이벤트를 전송합니다.
func (ctl *Controller) SendSSEMessageAll(ctx *gin.Context) {

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg sse.Message
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}
	// 이벤트 브로드캐스트
	ctl.SseService.Broadcast(msg)

	// 성공 응답
	ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": "Msg sent to all user"})
}

// SendEventToUser는 특정 사용자에게 이벤트를 전송합니다.
func (ctl *Controller) SendSSEMessageToUser(ctx *gin.Context) {
	userId := ctx.Param("id")

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg sse.Message
	err := ctx.ShouldBindJSON(&msg)
	if err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	// 사용자 세션 조회
	session := ctl.SseService.GetSessionByUserId(userId)
	if session == nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	// 이벤트 브로드캐스트
	session.Broadcast(msg)

	// 성공 응답
	ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": fmt.Sprintf("Event sent to user %s", userId)})
}
