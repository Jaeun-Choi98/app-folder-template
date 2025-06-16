package service

import (
	sse "pjt/internal/service/sse-service"

	"github.com/gin-gonic/gin"
)

type APIServcieInterface interface {
	Test() string
	Close() error
}

type SSEServiceInterface interface {
	Close()
	NewSession(userId string) *sse.UserSession
	NewSSEClient(clientId, userId string, ctx *gin.Context) (*sse.SSEClient, error)
	GetSessionById(sessionId string) *sse.UserSession
	GetSessionByUserId(userId string) *sse.UserSession
	Broadcast(msg sse.Message)
	RemoveSession(userId string)
}
