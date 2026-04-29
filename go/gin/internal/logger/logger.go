package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var customLogger *CustomLogger
var maxAgeDays = -3
var rootDir = `log`

type CustomLogger struct {
	prefix    string
	curDay    int
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	ticker    *time.Ticker
	info      *log.Logger
	debug     *log.Logger
	infoFile  *os.File
	debugFile *os.File
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
	outfilepath := getFilepath("info", now)
	os.MkdirAll(filepath.Dir(outfilepath), 0755)
	infofile, err := os.OpenFile(outfilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	info := log.New(infofile, prefix, log.Ldate|log.Ltime)

	l.infoFile = infofile
	l.info = info

	outfilepath = getFilepath("debug", now)
	os.MkdirAll(filepath.Dir(outfilepath), 0755)
	debugFile, err := os.OpenFile(outfilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	debug := log.New(debugFile, prefix, log.Ldate|log.Ltime)

	l.debugFile = debugFile
	l.debug = debug

	l.curDay = now.Day()
	return nil
}

func getFilepath(name string, now time.Time) string {
	dir := filepath.Join(rootDir,
		fmt.Sprintf("%v", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%v", now.Day()),
	)

	return filepath.Join(dir, name)
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

func Infoln(v ...any) {
	log.Println(v...)
	customLogger.info.Println(v...)
}

func Infof(format string, v ...any) {
	log.Printf(format, v...)
	customLogger.info.Printf(format, v...)
}

func Debugln(v ...any) {
	log.Println(v...)
	customLogger.debug.Println(v...)
}

func Debugf(format string, v ...any) {
	log.Printf(format, v...)
	customLogger.debug.Printf(format, v...)
}

func Close() error {
	var terr error
	if err := customLogger.infoFile.Close(); err != nil {
		terr = fmt.Errorf("info file close err: %v", err)
	}
	if err := customLogger.debugFile.Close(); err != nil {
		terr = fmt.Errorf("debug file close err: %v", err)
	}

	return terr
}

func StartCleaning() {
	customLogger.wg.Add(1)
	defer customLogger.wg.Done()
	cleanOldLogs()

	for {
		select {
		case <-customLogger.ctx.Done():
			Infoln("[Log Cleanup] log cleanup goroutine terminated")
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

	root := filepath.Base(rootDir)

	deletedCount := searchLogFileAndDelete(root, time.Now().AddDate(0, 0, maxAgeDays))

	Infof("[Log Cleanup] log cleanup completed: %d files deleted in total", deletedCount)
}

func searchLogFileAndDelete(parent string, cutoffTime time.Time) int {

	// debug
	// log.Println(parent)

	files, err := os.ReadDir(parent)
	delCnt := 0

	if err != nil {
		Infof("[Log Cleanup] failed to searchLogFileAndDelete:\n\t %v", err)
		return 0
	}

	for _, file := range files {

		if file.IsDir() {
			delCnt += searchLogFileAndDelete(filepath.Join(parent, file.Name()), cutoffTime)
			continue
		}

		info, err := file.Info()
		if err != nil {
			Infof("[Log Cleanup] failed to get file information:\n\t %v", err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {

			filePath := filepath.Join(parent, file.Name())
			Infof("[Log Cleanup] old log files deleted: %s (Deleted date: %s)",
				filePath, info.ModTime().Format("2006-01-02"))

			if err := os.Remove(filePath); err != nil {
				Infof("[Log Cleanup] failed to delete file:\n\t %v", err)
			} else {
				delCnt += 1
			}
		}
	}
	return delCnt
}

func Shutdown() {
	customLogger.cancel()
	customLogger.ticker.Stop()
	customLogger.wg.Wait()
	Close()
}
