package controller

import (
	"context"
	"fmt"
	"net/http"
	model "pjt/internal/model/sse"
	"pjt/internal/server/http-rest/http-utils/httperr"
	sse "pjt/internal/service/sse-service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleSSE는 클라이언트의 SSE 연결 요청을 처리합니다.
func (ctl *Controller) HandleSSEConnect(ctx *gin.Context) {
	userId := ctx.Param("id")

	// SSE 헤더 설정
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")

	// HTTP Flusher 확인
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		ctx.Error(httperr.INNER_ERROR.AddErrMsg(fmt.Errorf("streaming not supported")))
		return
	}

	// 클라이언트 ID 생성
	clientId := uuid.New().String()

	// 클라이언트 컨텍스트 생성
	clientCtx, cancel := context.WithCancel(ctx.Request.Context())

	// 새 SSE 클라이언트 생성
	client := &sse.SSEClient{
		ClientId: clientId,
		UserId:   userId,
		Writer:   ctx.Writer,
		Flusher:  flusher,
		Ctx:      clientCtx,
		Cancel:   cancel,
	}

	session := ctl.SseService.GetSessionByUserId(userId)
	if session == nil {
		session = ctl.SseService.NewSession(userId)
	}

	// 세션에 클라이언트 추가
	session.AddClient(clientId, client)

	// 연결 성공 이벤트 전송
	connectMsg := model.SSEMessage{
		Type:    "connect",
		Payload: map[string]string{"message": "Connected successfully", "client_id": clientId},
	}
	client.SendMessage(connectMsg)

	// 클라이언트 요청이 종료되면 정리
	// ctx.Request.Context()는 클라이언트가 연결을 끊으면 취소됩니다
	go func() {
		<-clientCtx.Done()
		session.RemoveClient(clientId)
		if session.Count() == 0 {
			ctl.SseService.RemoveSession(userId)
		}
	}()

	// 클라이언트가 연결을 끊을 때까지 연결 유지
	// 심박(heartbeat)을 30초마다 전송
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-clientCtx.Done():
			// 클라이언트 또는 서버에서 연결을 닫았음
			return
		case <-ticker.C:
			// 심박 전송 - 연결이 활성 상태인지 확인
			heartbeat := model.SSEMessage{
				Type:    "heartbeat",
				Payload: time.Now().Unix(),
			}
			err := client.SendMessage(heartbeat)
			if err != nil {
				// 오류 발생 시 연결 종료
				session.RemoveClient(clientId)
				return
			}
		}
	}
}

// SendEventToUser는 모든 사용자에게 이벤트를 전송합니다.
func (ctl *Controller) SendSSEMessageAll(ctx *gin.Context) {

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg model.SSEMessage
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		ctx.Error(httperr.INNER_ERROR.AddErrMsg(err))
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
	var msg model.SSEMessage
	err := ctx.ShouldBindJSON(&msg)
	if err != nil {
		ctx.Error(httperr.INNER_ERROR.AddErrMsg(err))
		return
	}

	if err != nil {
		ctx.Error(httperr.INNER_ERROR.AddErrMsg(err))
		return
	}

	// 사용자 세션 조회
	session := ctl.SseService.GetSessionByUserId(userId)
	if session == nil {
		ctx.Error(httperr.INNER_ERROR.AddErrMsg(err))
		return
	}

	// 이벤트 브로드캐스트
	session.Broadcast(msg)

	// 성공 응답
	ctx.JSON(http.StatusOK, map[string]interface{}{"success": true, "message": fmt.Sprintf("Event sent to user %s", userId)})
}
