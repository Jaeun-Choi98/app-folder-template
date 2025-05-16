package repository

import (
	"pjt/internal/config"
)

type DaoInterface interface {
	Test() string
}

func NewMyDB(config *config.Configuration) (DaoInterface, error) {
	return NewOralce(config)
}
