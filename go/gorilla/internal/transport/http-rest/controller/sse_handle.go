package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pjt/internal/logger"
	model "pjt/internal/model/sse"
	"pjt/internal/service"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// HandleSSE는 클라이언트의 SSE 연결 요청을 처리합니다.
func (c *Controller) HandleSSE(w http.ResponseWriter, r *http.Request) {

	sseManager, err := c.Service.GetSSEManager()
	if err != nil {
		http.Error(w, "SSE Manager is empty", http.StatusInternalServerError)
		return
	}

	// SSE 헤더 설정
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// HTTP Flusher 확인
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// 클라이언트 ID 생성
	clientId := uuid.New().String()

	// 클라이언트 컨텍스트 생성
	ctx, cancel := context.WithCancel(r.Context())

	// 새 SSE 클라이언트 생성
	client := &service.SSEClient{
		ClientId: clientId,
		Writer:   w,
		Flusher:  flusher,
		Ctx:      ctx,
		Cancel:   cancel,
	}

	userId := mux.Vars(r)["user_id"]
	// 사용자 세션 가져오기 또는 생성
	session := sseManager.GetSessionByUserId(userId)
	if session == nil {
		session = sseManager.NewSession(userId)
	}

	// 세션에 클라이언트 추가
	session.AddClient(clientId, client)

	// 연결 성공 이벤트 전송
	connectEvent := model.SSEMessage{
		Type:    "connect",
		Payload: map[string]string{"message": "Connected successfully", "client_id": clientId},
	}
	client.SendMessage(connectEvent)

	// 클라이언트 요청이 종료되면 정리
	// r.Context()는 클라이언트가 연결을 끊으면 취소됩니다
	go func() {
		<-r.Context().Done()
		logger.Printf("User[%s]'s client[%s] is closed", userId, clientId)
		session.RemoveClient(clientId)
		if session.Count() == 0 {
			sseManager.RemoveSession(userId)
		}
	}()

	// 클라이언트가 연결을 끊을 때까지 연결 유지
	// 심박(heartbeat)을 2초마다 전송
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
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
func (c *Controller) SendSSEMessageAll(w http.ResponseWriter, r *http.Request) {

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg model.SSEMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	sseManager, err := c.Service.GetSSEManager()
	if err != nil {
		http.Error(w, "SSE Manager is empty", http.StatusInternalServerError)
		return
	}

	// 이벤트 브로드캐스트
	sseManager.Broadcast(msg)

	// 성공 응답
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "message": "Msg sent to all user"}`)
}

// SendEventToUser는 특정 사용자에게 이벤트를 전송합니다.
func (c *Controller) SendSSEMessageToUser(w http.ResponseWriter, r *http.Request) {
	// URL에서 사용자 ID 추출
	vars := mux.Vars(r)
	userID := vars["userId"]

	// POST 요청 본문에서 이벤트 데이터 파싱
	var msg model.SSEMessage
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Invalid event data", http.StatusBadRequest)
		return
	}

	sseManager, err := c.Service.GetSSEManager()
	if err != nil {
		http.Error(w, "SSE Manager is empty", http.StatusInternalServerError)
		return
	}

	// 사용자 세션 조회
	session := sseManager.GetSessionByUserId(userID)
	if session == nil {
		http.Error(w, "User not connected", http.StatusNotFound)
		return
	}

	// 이벤트 브로드캐스트
	session.Broadcast(msg)

	// 성공 응답
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"success": true, "message": "Event sent to user %s"}`, userID)
}
