// Package logger 는 날짜별 디렉터리(log/YYYY/MM/D/)에 두 개의 파일을 남긴다.
//
//	app.log  - 애플리케이션 로그. 레벨(DEBUG < INFO < WARN < ERROR)로 거른다.
//	dump.log - TCP/UDP/Serial 프레임 hex 덤프. 레벨과 별개로 켜고 끈다.
//
// 레벨과 덤프 여부는 config 가 env.ini [LOG] 에서 읽은 값을 NewCustomLogger 에 넘겨 정하고,
// 실행 중에는 SetLevel/SetDump 로 바꾼다. 실행 중 변경은 파일에 저장하지 않는다 -
// 재기동하면 env.ini 값으로 돌아간다.
//
// config 를 읽는 동안처럼 로거가 만들어지기 전에 남긴 로그는 표준 log(표준에러)로만 나간다.
package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Level int32

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

var levelNames = [...]string{"DEBUG", "INFO", "WARN", "ERROR"}

func (lv Level) String() string {
	if lv < DEBUG || lv > ERROR {
		return fmt.Sprintf("Level(%d)", int32(lv))
	}
	return levelNames[lv]
}

// ParseLevel 은 "debug", "INFO" 같은 문자열을 레벨로 바꾼다. 대소문자는 무시한다.
func ParseLevel(s string) (Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(strings.TrimSpace(s), name) {
			return Level(i), nil
		}
	}
	return INFO, fmt.Errorf("unknown log level: %q", s)
}

var customLogger *CustomLogger
var maxAgeDays = -3
var rootDir = `log`

const (
	appFileName  = "app.log"
	dumpFileName = "dump.log"
	logFlags     = log.Ldate | log.Ltime | log.Lmicroseconds
)

type CustomLogger struct {
	prefix string
	level  atomic.Int32
	dump   atomic.Bool

	curDay int
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker

	appFile  *dailyFile
	dumpFile *dailyFile
	app      *log.Logger // 파일 + 콘솔
	dumper   *log.Logger // 파일만
}

// dailyFile 은 날짜가 바뀔 때 안쪽 파일만 바꿔 끼우는 writer 다.
// log.Logger 는 이 writer 를 계속 들고 있으므로 교체 중에도 다른 고루틴의 쓰기가 안전하다.
type dailyFile struct {
	mu   sync.Mutex
	name string
	f    *os.File
}

func (d *dailyFile) Write(p []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.f == nil {
		return len(p), nil
	}
	return d.f.Write(p)
}

func (d *dailyFile) open(now time.Time) error {
	path := getFilepath(d.name, now)
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	d.mu.Lock()
	old := d.f
	d.f = f
	d.mu.Unlock()

	if old != nil {
		old.Close()
	}
	return nil
}

func (d *dailyFile) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.f == nil {
		return nil
	}
	err := d.f.Close()
	d.f = nil
	return err
}

func NewCustomLogger(prefix string, level Level, dump bool) (*CustomLogger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	l := &CustomLogger{
		prefix:   prefix,
		ctx:      ctx,
		cancel:   cancel,
		ticker:   time.NewTicker(1 * time.Minute),
		appFile:  &dailyFile{name: appFileName},
		dumpFile: &dailyFile{name: dumpFileName},
	}
	l.level.Store(int32(level))
	l.dump.Store(dump)

	if err := l.openFiles(time.Now()); err != nil {
		cancel()
		l.ticker.Stop()
		return nil, err
	}

	l.app = log.New(&consoleTee{file: l.appFile}, prefix, logFlags)
	l.dumper = log.New(l.dumpFile, prefix, logFlags)
	return l, nil
}

// consoleTee 는 app.log 와 표준에러에 같이 쓴다 (기존 로거의 log.Printf 출력 유지).
type consoleTee struct {
	file *dailyFile
}

func (c *consoleTee) Write(p []byte) (int, error) {
	os.Stderr.Write(p)
	return c.file.Write(p)
}

func (l *CustomLogger) openFiles(now time.Time) error {
	if err := l.appFile.open(now); err != nil {
		log.Println(err)
		return err
	}
	if err := l.dumpFile.open(now); err != nil {
		log.Println(err)
		return err
	}
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

// ---- 레벨 / 덤프 설정 ----

func SetLevel(lv Level) { customLogger.level.Store(int32(lv)) }

// GetLevel 은 로거가 아직 없으면 INFO 를 돌려준다.
func GetLevel() Level {
	if customLogger == nil {
		return INFO
	}
	return Level(customLogger.level.Load())
}

func SetDump(on bool) { customLogger.dump.Store(on) }

// IsDumpEnabled 는 로거가 아직 없으면 false 다.
func IsDumpEnabled() bool { return customLogger != nil && customLogger.dump.Load() }

// Enabled 는 lv 레벨 로그가 기록되는지 알려준다.
// 인자를 만드는 비용이 큰 로그(hex 문자열 생성 등) 앞에서 쓴다.
func Enabled(lv Level) bool { return lv >= GetLevel() }

// ---- 애플리케이션 로그 ----

func output(lv Level, msg string) {
	if !Enabled(lv) {
		return
	}
	line := fmt.Sprintf("[%-5s] %s", lv, strings.TrimSuffix(msg, "\n"))
	if customLogger == nil {
		log.Output(3, line)
		return
	}
	customLogger.app.Output(3, line)
}

func Debugln(v ...any) { output(DEBUG, fmt.Sprintln(v...)) }

func Debugf(format string, v ...any) { output(DEBUG, fmt.Sprintf(format, v...)) }

func Infoln(v ...any) { output(INFO, fmt.Sprintln(v...)) }

func Infof(format string, v ...any) { output(INFO, fmt.Sprintf(format, v...)) }

func Warnln(v ...any) { output(WARN, fmt.Sprintln(v...)) }

func Warnf(format string, v ...any) { output(WARN, fmt.Sprintf(format, v...)) }

func Errorln(v ...any) { output(ERROR, fmt.Sprintln(v...)) }

func Errorf(format string, v ...any) { output(ERROR, fmt.Sprintf(format, v...)) }

// ---- 프레임 덤프 ----

type Direction string

const (
	Rx Direction = "Rx"
	Tx Direction = "Tx"
)

const dumpBytesPerLine = 16

// Dump 는 한 프레임을 dump.log 에 16바이트씩 끊어 기록한다.
//
//	2026/09/13 20:16:00.123456 [TCP] [Tx] peer=5001 op=0x21 len=12
//	  0000: 7E 7E 21 00 05 00 00 13 89 01 8C 7F
//
// link 는 "TCP"/"UDP"/"SERIAL", peer 는 client id 나 원격 주소처럼 상대를 구분할 값이다.
// frame 이 nil 이면 헤더만 남긴다 (파일 전송처럼 본문이 큰 프레임).
func Dump(link string, dir Direction, peer any, opCode byte, frame []byte) {
	if !IsDumpEnabled() {
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[%s] [%s] peer=%v op=0x%02X len=%d", link, dir, peer, opCode, len(frame))
	if frame == nil {
		sb.WriteString(" (body omitted)")
	}
	for i, b := range frame {
		if i%dumpBytesPerLine == 0 {
			fmt.Fprintf(&sb, "\n  %04X:", i)
		}
		fmt.Fprintf(&sb, " %02X", b)
	}
	customLogger.dumper.Output(2, sb.String())
}

// ---- 파일 교체 / 정리 ----

func Close() error {
	var terr error
	if err := customLogger.appFile.Close(); err != nil {
		terr = fmt.Errorf("app file close err: %v", err)
	}
	if err := customLogger.dumpFile.Close(); err != nil {
		terr = fmt.Errorf("dump file close err: %v", err)
	}

	return terr
}

// StartCleaning 은 자신의 고루틴을 띄우고 바로 반환한다.
// 1분 주기로 날짜 변경을 감지해 로그 파일을 교체하고 보존 기간이 지난 로그를 지운다.
func StartCleaning() {
	customLogger.wg.Add(1)
	go func() {
		defer customLogger.wg.Done()
		for cleanLoop() { // false = ctx가 끝나 정상 종료
		}
	}()
}

// cleanLoop 는 패닉이 나면 로그를 남기고 true 를 돌려 바깥 루프가 다시 돌게 한다.
func cleanLoop() (restart bool) {
	defer func() {
		if r := recover(); r != nil {
			Errorf("[Log Cleanup] panic recovered, restarting: %v\n%s", r, debug.Stack())
			restart = true
		}
	}()

	cleanOldLogs()

	for {
		select {
		case <-customLogger.ctx.Done():
			Infoln("[Log Cleanup] log cleanup goroutine terminated")
			return false
		case <-customLogger.ticker.C:
			now := time.Now()
			if customLogger.curDay != now.Day() {
				if err := customLogger.openFiles(now); err != nil {
					Errorf("[Log Cleanup] failed to rotate log files: %v", err)
				}
				cleanOldLogs()
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

	files, err := os.ReadDir(parent)
	delCnt := 0

	if err != nil {
		Errorf("[Log Cleanup] failed to searchLogFileAndDelete:\n\t %v", err)
		return 0
	}

	for _, file := range files {

		if file.IsDir() {
			delCnt += searchLogFileAndDelete(filepath.Join(parent, file.Name()), cutoffTime)
			continue
		}

		info, err := file.Info()
		if err != nil {
			Errorf("[Log Cleanup] failed to get file information:\n\t %v", err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {

			filePath := filepath.Join(parent, file.Name())
			Infof("[Log Cleanup] old log files deleted: %s (Deleted date: %s)",
				filePath, info.ModTime().Format("2006-01-02"))

			if err := os.Remove(filePath); err != nil {
				Errorf("[Log Cleanup] failed to delete file:\n\t %v", err)
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
