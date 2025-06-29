package service

import (
	"pjt/internal/service/sse-service/sse"

	"github.com/gin-gonic/gin"
)

/**
 * SessionManager는 모든 사용자 세션을 관리.
 * 애플리케이션 내에서 싱글톤으로 사용
 */
type SSEService struct {
	sessionManager *sse.SessionManager
}

// NewSessionManager는 새로운 세션 관리자를 생성
func NewSSEService() *SSEService {
	return &SSEService{
		sessionManager: sse.NewSessionManager(),
	}
}

// NewSession은 사용자 ID를 기반으로 새 세션을 생성
func (s *SSEService) NewSession(userId string) *sse.Session {
	return s.sessionManager.NewSession(userId)
}

func (s *SSEService) NewSSEClient(clientId, userId string, ctx *gin.Context) (*sse.SSEClient, error) {
	return sse.NewSSEClient(clientId, userId, ctx)
}

// GetSessionByID는 세션 ID로 세션을 조회
func (s *SSEService) GetSessionById(sessionId string) *sse.Session {
	return s.sessionManager.GetSessionById(sessionId)
}

// GetSessionByUserID는 사용자 ID로 세션을 조회
func (s *SSEService) GetSessionByUserId(userId string) *sse.Session {
	return s.sessionManager.GetSessionByUserId(userId)
}

// RemoveSession은 세션을 제거
func (s *SSEService) RemoveSession(userId string) {
	s.sessionManager.RemoveSession(userId)
}

func (s *SSEService) Broadcast(msg sse.Message) {
	s.sessionManager.Broadcast(msg)
}

// Close는 모든 세션을 정리
func (s *SSEService) Close() {
	s.sessionManager.Close()
}
