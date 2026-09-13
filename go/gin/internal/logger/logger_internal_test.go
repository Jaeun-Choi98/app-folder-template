package logger

import (
	"os"
	"strings"
	"testing"
	"time"
)

func setupTestLogger(t *testing.T, level Level, dump bool) {
	t.Helper()
	rootDir = t.TempDir()
	customLogger = nil
	l, err := NewCustomLogger("", level, dump)
	if err != nil {
		t.Fatal(err)
	}
	SetLogger(l)
	t.Cleanup(func() {
		l.ticker.Stop()
		Close()
		customLogger = nil
	})
}

func readLog(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(getFilepath(name, time.Now()))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// config 에서 넘긴 값이 생성 시점부터 적용되는지 확인한다.
func TestConstructorAppliesConfig(t *testing.T) {
	setupTestLogger(t, DEBUG, false)
	if GetLevel() != DEBUG {
		t.Errorf("level = %s, want DEBUG", GetLevel())
	}
	if IsDumpEnabled() {
		t.Error("dump should be disabled")
	}

	Debugf("first-debug")
	Dump("TCP", Tx, 1, 0x21, []byte{0x7E})
	if !strings.Contains(readLog(t, appFileName), "[DEBUG] first-debug") {
		t.Error("first DEBUG line was dropped")
	}
	if readLog(t, dumpFileName) != "" {
		t.Error("dump written while disabled")
	}
}

// 로거 생성 전(config 로딩 중)에 로그를 남겨도 패닉이 나지 않아야 한다.
func TestBeforeLoggerCreated(t *testing.T) {
	customLogger = nil
	Infof("before init %d", 1)
	Dump("TCP", Rx, 1, 0x21, []byte{0x7E})
	if GetLevel() != INFO || IsDumpEnabled() {
		t.Errorf("nil logger: level=%s dump=%v", GetLevel(), IsDumpEnabled())
	}
}

func TestLevelGating(t *testing.T) {
	setupTestLogger(t, INFO, true)

	Debugf("drop-debug")
	Infof("keep-info %d", 1)
	SetLevel(WARN)
	Infoln("drop-info")
	Warnf("keep-warn")
	Errorln("keep-error")

	got := readLog(t, appFileName)
	for _, s := range []string{"[INFO ] keep-info 1", "[WARN ] keep-warn", "[ERROR] keep-error"} {
		if !strings.Contains(got, s) {
			t.Errorf("app.log missing %q\n%s", s, got)
		}
	}
	for _, s := range []string{"drop-debug", "drop-info"} {
		if strings.Contains(got, s) {
			t.Errorf("app.log should not contain %q", s)
		}
	}
	if strings.Count(got, "\n") != 3 {
		t.Errorf("want 3 lines, got:\n%s", got)
	}
}

func TestDump(t *testing.T) {
	setupTestLogger(t, ERROR, true) // 덤프는 레벨과 무관해야 한다

	frame := make([]byte, 18)
	for i := range frame {
		frame[i] = byte(i)
	}
	Dump("TCP", Tx, 5001, 0x21, frame)
	Dump("TCP", Rx, 5001, 0x1D, nil)
	SetDump(false)
	Dump("TCP", Tx, 5001, 0x22, frame)

	got := readLog(t, dumpFileName)
	want := []string{
		"[TCP] [Tx] peer=5001 op=0x21 len=18\n  0000: 00 01 02 03 04 05 06 07 08 09 0A 0B 0C 0D 0E 0F\n  0010: 10 11\n",
		"[TCP] [Rx] peer=5001 op=0x1D len=0 (body omitted)\n",
	}
	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Errorf("dump.log missing %q\n%s", s, got)
		}
	}
	if strings.Contains(got, "op=0x22") {
		t.Error("dump written while disabled")
	}
	if strings.Contains(readLog(t, appFileName), "op=0x21") {
		t.Error("dump leaked into app.log")
	}
}

func TestParseLevel(t *testing.T) {
	for in, want := range map[string]Level{"debug": DEBUG, " INFO ": INFO, "Warn": WARN, "ERROR": ERROR} {
		if got, err := ParseLevel(in); err != nil || got != want {
			t.Errorf("ParseLevel(%q) = %v, %v", in, got, err)
		}
	}
	if _, err := ParseLevel("TRACE"); err == nil {
		t.Error("want error for unknown level")
	}
}
