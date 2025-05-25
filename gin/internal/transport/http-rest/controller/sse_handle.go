package controller

/*
// HandleSSE는 클라이언트의 SSE 연결 요청을 처리합니다.
func (ctr *Controller) HandleSSE(w http.ResponseWriter, r *http.Request) {

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
	clientID := uuid.New().String()

	// 클라이언트 컨텍스트 생성
	ctx, cancel := context.WithCancel(r.Context())

	// 새 SSE 클라이언트 생성
	client := &service.SSEClient{
		ClientId: clientID,
		Writer:   w,
		Flusher:  flusher,
		Ctx:      ctx,
		Cancel:   cancel,
	}

	// 사용자 세션 가져오기 또는 생성
	session := c.Service.GetSessionByUserID(userID)
	if session == nil {
		session = c.sessionManager.NewSession(userID)
	}

	// 세션에 클라이언트 추가
	session.AddClient(clientID, client)

	// 연결 성공 이벤트 전송
	connectEvent := model.Event{
		EventType: "connect",
		Data:      map[string]string{"message": "Connected successfully", "client_id": clientID},
	}
	client.SendEvent(connectEvent)

	// 클라이언트 요청이 종료되면 정리
	// 주: r.Context()는 클라이언트가 연결을 끊으면 취소됩니다
	go func() {
		<-r.Context().Done()
		session.RemoveClient(clientID)
	}()

	// 클라이언트가 연결을 끊을 때까지 연결 유지
	// 심박(heartbeat)을 30초마다 전송
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 클라이언트 또는 서버에서 연결을 닫았음
			return
		case <-ticker.C:
			// 심박 전송 - 연결이 활성 상태인지 확인
			heartbeat := model.Event{
				EventType: "heartbeat",
				Data:      time.Now().Unix(),
			}
			err := client.SendEvent(heartbeat)
			if err != nil {
				// 오류 발생 시 연결 종료
				session.RemoveClient(clientID)
				return
			}
		}
	}
}
*/
