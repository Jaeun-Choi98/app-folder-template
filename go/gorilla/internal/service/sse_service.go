package service

import (
	"context"
	"fmt"
	"net/http"
	"pjt/internal/logger"
	model "pjt/internal/model/sse"
	"sync"
	"time"
)

/**
 * SSEClient는 SSE 연결을 유지하는 개별 클라이언트를 나타냄
 * 각 클라이언트는 고유 ID와 응답 Writer를 가짐
 */
type SSEClient struct {
	ClientId string
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Ctx      context.Context
	Cancel   context.CancelFunc
}

// SendEvent는 SSE 클라이언트에게 이벤트를 전송
func (c *SSEClient) SendMessage(event model.SSEMessage) error {
	// 컨텍스트가 취소되었는지 확인
	select {
	case <-c.Ctx.Done():
		err := fmt.Errorf("client connection closed")
		logger.Printf("Failed to send message:\n\t%v", err)
		return err
	default:
		// SSE 형식으로 이벤트 전송
		fmt.Fprintf(c.Writer, "event: %s\n", event.Type)
		fmt.Fprintf(c.Writer, "data: %v\n\n", event.Payload)
		c.Flusher.Flush()
		return nil
	}
}

// Close는 클라이언트 연결을 종료
func (c *SSEClient) Close() {
	c.Cancel()
}

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
func (s *UserSession) AddClient(clientID string, client *SSEClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[clientID] = client
}

// RemoveClient는 세션에서 SSE 클라이언트를 제거
func (s *UserSession) RemoveClient(clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if client, exists := s.clients[clientID]; exists {
		client.Close()
		delete(s.clients, clientID)
	}
}

func (s *UserSession) Count() int {
	return len(s.clients)
}

// Broadcast는 세션의 모든 클라이언트에게 이벤트를 전송
func (s *UserSession) Broadcast(msg model.SSEMessage) {
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

// Close는 세션의 모든 클라이언트 연결을 종료
func (s *UserSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, client := range s.clients {
		if client != nil {
			client.Close()
		}
	}
	s.clients = make(map[string]*SSEClient)
}

/**
 * SessionManager는 모든 사용자 세션을 관리.
 * 애플리케이션 내에서 싱글톤으로 사용
 */
type SSESessionManager struct {
	sessions  map[string]*UserSession // 세션 ID를 키로 하는 맵
	userIndex map[string]string       // 사용자 ID를 키로, 세션 ID를 값으로 하는 맵
	mu        *sync.RWMutex
}

// NewSessionManager는 새로운 세션 관리자를 생성
func NewSessionManager() *SSESessionManager {
	return &SSESessionManager{
		sessions:  make(map[string]*UserSession),
		userIndex: make(map[string]string),
		mu:        &sync.RWMutex{},
	}
}

// NewSession은 사용자 ID를 기반으로 새 세션을 생성
func (m *SSESessionManager) NewSession(userId string) *UserSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 이미 존재하는 세션이 있으면 반환
	if sessionId, exists := m.userIndex[userId]; exists {
		return m.sessions[sessionId]
	}

	// 새 세션 ID 생성
	sessionId := fmt.Sprintf("sess_%s_%d", userId, time.Now().UnixNano())
	session := NewUserSession(sessionId, userId)

	m.sessions[sessionId] = session
	m.userIndex[userId] = sessionId

	return session
}

// GetSessionByID는 세션 ID로 세션을 조회
func (m *SSESessionManager) GetSessionById(sessionId string) *UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[sessionId]
}

// GetSessionByUserID는 사용자 ID로 세션을 조회
func (m *SSESessionManager) GetSessionByUserId(userId string) *UserSession {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if sessionID, exists := m.userIndex[userId]; exists {
		return m.sessions[sessionID]
	}
	return nil
}

// Broadcast는 세션의 모든 클라이언트에게 이벤트를 전송
func (m *SSESessionManager) Broadcast(msg model.SSEMessage) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, session := range m.sessions {
		session.Broadcast(msg)
	}
}

// RemoveSession은 세션을 제거
func (m *SSESessionManager) RemoveSession(userId string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sessionId, exists := m.userIndex[userId]; exists {
		// 연결 종료
		if session, exists := m.sessions[userId]; exists {
			session.Close()
		}

		// 인덱스에서 제거
		delete(m.userIndex, userId)
		// 세션 맵에서 제거
		delete(m.sessions, sessionId)
		logger.Printf("User[%s]'s session[%s] is closed", userId, sessionId)
	}
}

// Shutdown은 모든 세션을 정리
func (m *SSESessionManager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 모든 세션 종료
	for _, session := range m.sessions {
		session.Close()
	}

	// 맵 초기화
	m.sessions = make(map[string]*UserSession)
	m.userIndex = make(map[string]string)
}
