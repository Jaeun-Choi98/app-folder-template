package repository

import (
	"database/sql"
	"pjt/internal/config"
)

type DaoInterface interface {
	Test() string
}

type MyDB struct {
	db *sql.DB
}

func NewMyDB(config *config.Configuration) DaoInterface {
	return &MyDB{}
}

func (m *MyDB) Test() string {
	return "hello world"
}
