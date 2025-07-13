package service

import (
	"pjt/internal/service/sse-service/sse"

	"github.com/gin-gonic/gin"
)

type APIServcieInterface interface {
	Test() string
	Close() error
}

type SSEServiceInterface interface {
	Close()
	NewSession(userId string) *sse.Session
	NewSSEClient(clientId uint32, userId string, ctx *gin.Context) (*sse.SSEClient, error)
	GetSessionById(sessionId string) *sse.Session
	GetSessionByUserId(userId string) *sse.Session
	Broadcast(msg sse.Message)
	RemoveSession(userId string)
}

type TCPServiceInterface interface {
	Handle0x010() error
}
