package dbhandler

import (
	"pjt/internal/config"
)

type DBHandlerInterface interface {
	Test() string
	Close() error
	StartDBHeartbeat()
	StopDBHeartbeat()
}

func NewMyDB(config *config.Configuration) (DBHandlerInterface, error) {
	return NewOralce(config)
}
