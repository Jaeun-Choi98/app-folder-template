package service

import (
	ssemodel "pjt/internal/model/sse"
	sse "pjt/internal/service/sse-service"
)

type APIServcieInterface interface {
	Test() string
	Close() error
}

type SSEServiceInterface interface {
	Close()
	NewSession(userId string) *sse.UserSession
	GetSessionById(sessionId string) *sse.UserSession
	GetSessionByUserId(userId string) *sse.UserSession
	Broadcast(msg ssemodel.SSEMessage)
	RemoveSession(userId string)
}
