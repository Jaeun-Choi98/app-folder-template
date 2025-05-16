package repository

import (
	"database/sql"
	"fmt"
	"log"
	"pjt/internal/config"
	//_ "github.com/godror/godror"
)

type Oracle struct {
	config *config.Configuration
	db     *sql.DB
}

func NewOralce(config *config.Configuration) (DaoInterface, error) {
	connInfo := fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, config.User, config.Passwd, config.Conn)
	db, err := sql.Open("godror", connInfo)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &Oracle{
		db:     db,
		config: config,
	}, nil
}

func (o *Oracle) Test() string {
	return "hello world"
}

func (o *Oracle) Ping() error {
	return o.db.Ping()
}

func (o Oracle) Close() error {
	return o.db.Close()
}
