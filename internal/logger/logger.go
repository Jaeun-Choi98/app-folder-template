package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type CustomLogger struct {
	Logger *log.Logger
}

func NewCustomLogger(prefix string) (*CustomLogger, error) {

	filepath := getFilepath()
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	logger := log.New(io.MultiWriter(os.Stderr, file), prefix, log.Ldate|log.Ltime|log.Llongfile)

	return &CustomLogger{
		Logger: logger,
	}, nil
}

func (c *CustomLogger) Println(str string) {
	c.Logger.Println(str)
}

func (c *CustomLogger) Printf(format string, v ...any) {
	c.Logger.Printf(format, v...)
}

func getFilepath() string {
	now := time.Now()
	dir := filepath.Base("log")
	fileName := fmt.Sprintf("%d_%02d_%02d", now.Year(), now.Month(), now.Day())
	return filepath.Join(dir, fileName)
}
