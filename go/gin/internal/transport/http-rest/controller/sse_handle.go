package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"pjt/internal/event"
	"pjt/internal/logger"

	"pjt/internal/transport/http-rest/http-utils/httperr"
	"pjt/internal/transport/http-rest/response"

	"github.com/Jaeun-Choi98/modules/eventbus"
	"github.com/Jaeun-Choi98/modules/sse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CustomSSEMsg struct {
	Type    string
	Payload any
}

func (m *CustomSSEMsg) GetEvent() any {
	return m.Type
}

func (m *CustomSSEMsg) GetData() any {
	return m.Payload
}

func (m *CustomSSEMsg) GetId() any {
	return nil
}

func (m *CustomSSEMsg) GetRetry() any {
	return nil
}

/**
 * 클라이언트의 SSE 연결 요청을 처리 하기 위한 핸들러
 * 사용자의 ID가 없다면, ClientId를 UserId로 사용.
 *
 * 사용자ID가 없다면:
 * -> 특정 사용자에게만 데이터를 전송하는 /login/sse-send/:id API를 사용할 수 없음.
 * -> 하나의 세션에 하나의 클라이언트만 매칭 됨.
 */
func (ctl *Controller) HandleSSEConnect(ctx *gin.Context) {

	clientId := uuid.New().ID()
	userId := uint32(ctx.GetInt("id"))
	if userId == 0 {
		userId = clientId
	}

	client, err := sse.NewSSEClient[uint32, uint32](clientId, userId, ctx)
	if err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	session := ctl.SseManager.GetSessionByUserId(userId)
	if session == nil {
		session = ctl.SseManager.NewSession(userId)
	}

	// 세션에 클라이언트 추가
	session.AddClient(clientId, client)

	// 연결 성공 이벤트 전송
	connectMsg := &CustomSSEMsg{
		Type:    "connect",
		Payload: map[string]any{"message": "Connected successfully", "client_id": clientId},
	}
	client.SendMessage(connectMsg)

	unsubscribe := eventbus.Subscribe[*event.SseEvent](event.GetEventBus(), "sse.event",
		func(se *event.SseEvent) (any, error) {
			payloadByte, err := json.Marshal(se.Data)
			var msg *CustomSSEMsg
			if err != nil {
				logger.Infoln("[SSE] failed to serialize event payload to json")
				msg = &CustomSSEMsg{
					Type:    se.Event,
					Payload: se.Data,
				}
			} else {
				msg = &CustomSSEMsg{
					Type:    se.Event,
					Payload: string(payloadByte),
				}
			}
			if err := client.SendMessage(msg); err != nil {
				// 오류 발생 시 연결 종료
				session.RemoveClient(clientId)
				return nil, err
			}
			return nil, nil
		})

	go func() {
		<-client.Ctx.Done()
		session.RemoveClient(clientId)
		unsubscribe()
		if session.Count() == 0 {
			ctl.SseManager.RemoveSession(userId)
		}
	}()

	// 클라이언트가 연결을 끊을 때까지 연결 유지
	select {
	case <-client.Ctx.Done():
		// 클라이언트 또는 서버에서 연결을 닫았음
		return
	}
}

// SendEventToUser는 모든 사용자에게 이벤트를 전송합니다.
func (ctl *Controller) SendSSEMessageAll(ctx *gin.Context) {

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg *CustomSSEMsg
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}
	// 이벤트 브로드캐스트
	ctl.SseManager.Broadcast(msg)

	// 성공 응답
	ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": "Msg sent to all user"})
}

// SendEventToUser는 특정 사용자에게 이벤트를 전송합니다.
func (ctl *Controller) SendSSEMessageToUser(ctx *gin.Context) {
	userId := ctx.Param("id")
	id, _ := strconv.Atoi(userId)

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg *CustomSSEMsg
	err := ctx.ShouldBindJSON(&msg)
	if err != nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	// 사용자 세션 조회
	session := ctl.SseManager.GetSessionByUserId(uint32(id))
	if session == nil {
		ctx.Error(httperr.INNER_ERROR.Add(err, response.FAIL))
		return
	}

	// 이벤트 브로드캐스트
	session.Broadcast(msg)

	// 성공 응답
	ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": fmt.Sprintf("Message sent to user %s", userId)})
}
