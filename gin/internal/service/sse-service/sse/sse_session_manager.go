package sse

import (
	"fmt"
	"pjt/internal/logger"
	"sync"
	"time"
)

type SessionManager struct {
	sessions  map[string]*Session // 세션 ID를 키로 하는 맵
	userIndex map[string]string   // 사용자 ID를 키로, 세션 ID를 값으로 하는 맵
	mu        sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:  make(map[string]*Session),
		userIndex: make(map[string]string),
	}
}

// NewSession은 사용자 ID를 기반으로 새 세션을 생성
func (s *SessionManager) NewSession(userId string) *Session {
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
func (s *SessionManager) GetSessionById(sessionId string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[sessionId]
}

// GetSessionByUserID는 사용자 ID로 세션을 조회
func (s *SessionManager) GetSessionByUserId(userId string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sessionID, exists := s.userIndex[userId]; exists {
		return s.sessions[sessionID]
	}
	return nil
}

// RemoveSession은 세션을 제거
func (s *SessionManager) RemoveSession(userId string) {
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

func (s *SessionManager) Broadcast(msg Message) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, session := range s.sessions {
		session.Broadcast(msg)
	}
}

// Close는 모든 세션을 정리
func (s *SessionManager) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 모든 세션 종료
	for _, session := range s.sessions {
		session.Close()
	}

	// 맵 초기화
	s.sessions = make(map[string]*Session)
	s.userIndex = make(map[string]string)
}
