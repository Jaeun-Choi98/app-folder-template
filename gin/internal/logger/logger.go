package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// global variable (package-level variable x)
var customLogger *CustomLogger
var maxAgeDays = 0

type CustomLogger struct {
	prefix  string
	curDay  int
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
	ticker  *time.Ticker
	logger  *log.Logger
	logFile *os.File
}

func NewCustomLogger(prefix string) (*CustomLogger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	l := &CustomLogger{
		prefix: prefix,
		ctx:    ctx,
		cancel: cancel,
		ticker: time.NewTicker(1 * time.Minute),
	}
	if err := l.setLogfileAndLogger(l.prefix); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *CustomLogger) setLogfileAndLogger(prefix string) error {
	now := time.Now()
	filepath := getFilepath(now)
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	logger := log.New(io.MultiWriter(os.Stderr, file), prefix, log.Ldate|log.Ltime)
	l.logFile = file
	l.logger = logger
	l.curDay = now.Day()
	return nil
}

func getFilepath(now time.Time) string {
	dir := filepath.Base("log")
	fileName := fmt.Sprintf("%d_%02d_%02d", now.Year(), now.Month(), now.Day())
	return filepath.Join(dir, fileName)
}

func SetLogger(l *CustomLogger) {
	if customLogger != nil {
		return
	}
	customLogger = l
}

// func GetLogger() *CustomLogger {
// 	return customLogger
// }

func Println(v ...any) {
	customLogger.logger.Println(v...)
}

func Printf(format string, v ...any) {
	customLogger.logger.Printf(format, v...)
}

func Close() error {
	return customLogger.logFile.Close()
}

/**
 * 일정 시간이 지난 로그 파일은 삭제하기 위한 고루틴 함수
 * Log prefix: [Log Cleanup]
 */
func StartCleaning() {
	customLogger.wg.Add(1)
	defer customLogger.wg.Done()
	cleanOldLogs()

	for {
		select {
		case <-customLogger.ctx.Done():
			Println("[Log Cleanup] Log cleanup goroutine terminated")
			return
		case <-customLogger.ticker.C:
			now := time.Now()
			if customLogger.curDay != now.Day() {
				Close()
				customLogger.setLogfileAndLogger(customLogger.prefix)
				cleanOldLogs()
				customLogger.curDay = now.Day()
			}

		}
	}
}

func cleanOldLogs() {
	cutoffTime := time.Now().AddDate(0, 0, maxAgeDays)

	dir := filepath.Base("log")

	files, err := os.ReadDir(dir)
	if err != nil {
		Printf("[Log Cleanup] Failed to read log directory:\n\t %v", err)
		return
	}

	var deletedCount int

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			Printf("[Log Cleanup] Failed to get file information:\n\t %v", err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {

			filePath := filepath.Join(dir, file.Name())
			Printf("[Log Cleanup] Old log files deleted: %s (Deleted date: %s)",
				filePath, info.ModTime().Format("2006-01-02"))

			if err := os.Remove(filePath); err != nil {
				Printf("[Log Cleanup] Failed to delete file:\n\t %v", err)
			} else {
				deletedCount++
			}
		}
	}

	Printf("[Log Cleanup] Log cleanup completed: %d files deleted in total", deletedCount)
}

func Shutdown() {
	customLogger.cancel()
	customLogger.ticker.Stop()
	customLogger.wg.Wait()
	Close()
}
