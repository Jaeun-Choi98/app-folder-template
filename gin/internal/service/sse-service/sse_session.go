package sse

import (
	"pjt/internal/logger"
	"sync"
)

/**
 * UserSession은 특정 사용자에 대한 모든 SSE 클라이언트 연결을 관리
 * 한 사용자는 여러 클라이언트(브라우저 탭, 기기 등)에서 연결
 */
type UserSession struct {
	sessionId string
	userId    string
	clients   map[string]*SSEClient
	mu        *sync.RWMutex
}

// NewUserSession은 새로운 사용자 세션을 생성
func NewUserSession(sessionId, userId string) *UserSession {
	return &UserSession{
		sessionId: sessionId,
		userId:    userId,
		clients:   make(map[string]*SSEClient),
		mu:        &sync.RWMutex{},
	}
}

// AddClient는 새 SSE 클라이언트를 세션에 추가
func (s *UserSession) AddClient(clientId string, client *SSEClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[clientId] = client
}

// RemoveClient는 세션에서 SSE 클라이언트를 제거
func (s *UserSession) RemoveClient(clientId string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client, exists := s.clients[clientId]; exists {
		client.Close()
		logger.Printf("Closed client[userId: %s, clientId: %s]", s.userId, clientId)
		delete(s.clients, clientId)
	}
}

// Broadcast는 세션의 모든 클라이언트에게 이벤트를 전송
func (s *UserSession) Broadcast(msg Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for clientId, client := range s.clients {
		err := client.SendMessage(msg)
		if err != nil {
			// 오류 발생 시 클라이언트 제거
			go s.RemoveClient(clientId)
		}
	}
}

func (s *UserSession) Count() int {
	return len(s.clients)
}

// Close는 세션의 모든 클라이언트 연결을 종료
func (s *UserSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		client.Close()
	}
	s.clients = make(map[string]*SSEClient)
}
