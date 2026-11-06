package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"pjt/internal/app"
	"pjt/internal/config"
	"pjt/internal/container"
	"syscall"
	"time"

	"github.com/Jaeun-Choi98/modules/logger"
)

const restartDelay = 3 * time.Second

func main() {

	cfg, err := config.NewConfiguration()
	if err != nil {
		log.Printf("[Main] failed to create new configuration, err: %v", err)
		os.Exit(1)
	}

	level, levelErr := logger.ParseLevel(cfg.LogLevel)
	// 패키지 기본 로거로 등록 ( log/YYYY/MM/D/app.log, dump.log, 오늘 + 지난 3일 보존 )
	if _, err := logger.Init(logger.Options{
		Dir:        logger.DefaultDir,
		Level:      level,
		Dump:       cfg.LogDump,
		MaxAgeDays: logger.DefaultMaxAgeDays,
		Console:    os.Stderr,
	}); err != nil {
		log.Printf("[Main] failed to init logger, err: %v", err)
		os.Exit(1)
	}
	if levelErr != nil {
		logger.Warnf("[Main] %v, fallback to %s", levelErr, level)
	}
	logger.Infof("[Main] log level: %s, dump: %v", level, cfg.LogDump)

	switch cfg.DeployMode {
	case config.TEST:
		os.Exit(Test(cfg))
	case config.PRODUCT:
		os.Exit(Product(cfg))
	default:
		log.Printf("[Main] undefined deploy stage, deploy mode: %v", cfg.DeployMode)
		os.Exit(1)
	}
}

func run(cfg *config.Configuration) (cause error, err error) {
	if err := container.NewContainer(cfg); err != nil {
		return nil, err
	}
	defer container.DelContainer()

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	a := app.NewApplication(cancel)
	a.Start()
	<-ctx.Done()

	return context.Cause(ctx), nil
}

func Product(cfg *config.Configuration) int {
	for {
		cause, err := run(cfg)
		if err != nil {
			log.Printf("[Main] failed to create container, err: %v", err)
			return 1
		}

		if errors.Is(cause, app.ErrSignal) {
			log.Printf("[Main] main process terminated by user (%v)", cause)
			return 0
		}

		log.Printf("[Main] main process terminated abnormally (%v). restarting in %v.", cause, restartDelay)
		if !waitRestart(restartDelay) {
			log.Printf("[Main] received signal while waiting to restart. exit.")
			return 0
		}
	}
}

func Test(cfg *config.Configuration) int {
	cause, err := run(cfg)
	if err != nil {
		log.Printf("[Main] failed to create container, err: %v", err)
		return 1
	}

	if errors.Is(cause, app.ErrSignal) {
		log.Printf("[Main] main process terminated by user (%v)", cause)
		return 0
	}
	log.Printf("[Main] main process terminated abnormally (%v)", cause)
	return 1
}

func waitRestart(d time.Duration) bool {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-time.After(d):
		return true
	case <-ctx.Done():
		return false
	}
}
