package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"pjt/internal/config"
	"pjt/internal/logger"
	"pjt/internal/server"
	"runtime/debug"
	"syscall"
	"time"
)

const shutdownTimeout = 5 * time.Second

// run 이 돌려주는 종료 원인. main 은 errors.Is 로 구분해 종료 코드를 정한다.
//
//	errSignal - 사용자/운영자가 SIGINT·SIGTERM 으로 종료를 요청함 → 0
//	errServer - 서버가 스스로 멈춤 (포트 사용 중, 인증서 오류 등)  → 1
//	errPanic  - 서버 고루틴에서 panic 이 발생함                    → 1
//
// 요청 처리 중의 panic 은 gin.Recovery 가 잡아 500 으로 응답하므로 여기까지 오지 않는다.
var (
	errSignal = errors.New("terminated by signal")
	errServer = errors.New("terminated by server error")
	errPanic  = errors.New("terminated by panic")
)

// os.Exit 는 defer 를 실행하지 않으므로, 정리(defer)는 run 안에서 끝내고 종료 코드만 밖으로 넘긴다.
func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.NewConfiguration()
	if err != nil {
		logger.Errorf("[Main] failed to load configuration: %v", err)
		return 1
	}

	level, levelErr := logger.ParseLevel(cfg.LogLevel)
	l, err := logger.NewCustomLogger("", level)
	if err != nil {
		logger.Errorf("[Main] failed to create logger: %v", err)
		return 1
	}
	logger.SetLogger(l)
	if levelErr != nil {
		logger.Warnf("[Main] %v, fallback to %s", levelErr, level)
	}
	logger.StartCleaning()
	defer logger.Shutdown()

	cause := serve(cfg)
	if errors.Is(cause, errSignal) {
		logger.Infof("[Main] application terminated by user (%v)", cause)
		return 0
	}
	logger.Errorf("[Main] application terminated abnormally (%v)", cause)
	return 1
}

// serve 는 서버를 띄우고, 시그널·서버 오류·panic 중 먼저 일어난 것을 원인으로 정리한 뒤 돌려준다.
func serve(cfg *config.Configuration) (cause error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	srv := server.New(cfg)
	stopped := make(chan error, 1) // 서버 고루틴의 종료 원인
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorf("[Main] server goroutine panic: %v\n%s", r, debug.Stack())
				stopped <- fmt.Errorf("%w: %v", errPanic, r)
			}
		}()
		err := srv.Start()
		if errors.Is(err, http.ErrServerClosed) {
			stopped <- nil // Shutdown 으로 닫힘 (정상)
			return
		}
		stopped <- fmt.Errorf("%w: %v", errServer, err)
	}()
	logger.Infof("[Main] server listening on %s:%s", cfg.RestIp, cfg.RestPort)

	select {
	case sig := <-sigCh:
		cause = fmt.Errorf("%w: %v", errSignal, sig)
	case cause = <-stopped:
		// 서버가 먼저 멈춘 경우, 정리할 대상이 이미 없으므로 바로 돌려준다.
		return cause
	}

	logger.Infof("[Main] shutting down... (cause: %v)", cause)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("[Main] shutdown error: %v", err)
	}

	// 서버 고루틴이 끝날 때까지 기다린다. 정리 중 panic 이 났다면 그것도 로그로 남긴다.
	select {
	case err := <-stopped:
		if err != nil {
			logger.Errorf("[Main] server stopped with error during shutdown: %v", err)
		}
	case <-ctx.Done():
		logger.Errorf("[Main] server goroutine did not stop within %v", shutdownTimeout)
	}
	return cause
}
