package service

import (
	"fmt"
	"pjt/internal/logger"
	ssemodel "pjt/internal/model/sse"
	"sync"
	"time"
)

/**
 * SessionManager는 모든 사용자 세션을 관리.
 * 애플리케이션 내에서 싱글톤으로 사용
 */
type SSEService struct {
	sessions  map[string]*UserSession // 세션 ID를 키로 하는 맵
	userIndex map[string]string       // 사용자 ID를 키로, 세션 ID를 값으로 하는 맵
	mu        sync.RWMutex
}

// NewSessionManager는 새로운 세션 관리자를 생성
func NewSSEService() *SSEService {
	return &SSEService{
		sessions:  make(map[string]*UserSession),
		userIndex: make(map[string]string),
	}
}

// NewSession은 사용자 ID를 기반으로 새 세션을 생성
func (s *SSEService) NewSession(userId string) *UserSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 이미 존재하는 세션이 있으면 반환
	if sessionId, exists := s.userIndex[userId]; exists {
		return s.sessions[sessionId]
	}

	// 새 세션 ID 생성
	sessionId := fmt.Sprintf("sess_%s_%d", userId, time.Now().UnixNano())
	session := NewUserSession(sessionId, userId)

	s.sessions[sessionId] = session
	s.userIndex[userId] = sessionId

	return session
}

// GetSessionByID는 세션 ID로 세션을 조회
func (s *SSEService) GetSessionById(sessionId string) *UserSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionId]
}

// GetSessionByUserID는 사용자 ID로 세션을 조회
func (s *SSEService) GetSessionByUserId(userId string) *UserSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sessionID, exists := s.userIndex[userId]; exists {
		return s.sessions[sessionID]
	}
	return nil
}

// RemoveSession은 세션을 제거
func (s *SSEService) RemoveSession(userId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sessionId, exists := s.userIndex[userId]; exists {

		// 연결 종료
		if session, exists := s.sessions[sessionId]; exists {
			session.Close()
		}

		logger.Printf("Closed session[userId:%s, sessionId:%s]", userId, sessionId)
		// 인덱스에서 제거
		delete(s.userIndex, userId)

		// 세션 맵에서 제거
		delete(s.sessions, sessionId)
	}
}

func (s *SSEService) Broadcast(msg ssemodel.SSEMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.sessions {
		session.Broadcast(msg)
	}
}

// Close는 모든 세션을 정리
func (s *SSEService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 모든 세션 종료
	for _, session := range s.sessions {
		session.Close()
	}

	// 맵 초기화
	s.sessions = make(map[string]*UserSession)
	s.userIndex = make(map[string]string)
}
