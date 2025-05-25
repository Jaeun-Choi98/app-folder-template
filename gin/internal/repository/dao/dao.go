package repository

import (
	"pjt/internal/config"
)

type DaoInterface interface {
	Test() string
	Close() error
	StartDBHeartbeat()
	StopDBHeartbeat()
}

func NewMyDB(config *config.Configuration) (DaoInterface, error) {
	return NewOralce(config)
}
