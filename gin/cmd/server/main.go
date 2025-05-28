package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"pjt/internal/app"
	"pjt/internal/container"
	"pjt/internal/logger"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

/**
 * AppManager는 애플리케이션의 생명주기와 상태를 관리합니다.
 * 패닉 복구, 하트비트 모니터링, 자동 재시작 기능을 제공합니다.
 */
type AppManager struct {
	container       *container.Container
	app             *app.Application
	ctx             context.Context
	cancel          context.CancelFunc
	isRunning       bool
	lastHeartbeat   time.Time
	restartCount    int
	maxRestarts     int
	heartbeatTicker *time.Ticker
	watchdogTicker  *time.Ticker
	mu              sync.RWMutex
	wg              sync.WaitGroup
}

// NewAppManager는 새로운 AppManager 인스턴스를 생성합니다
func NewAppManager(maxRestarts int) *AppManager {
	return &AppManager{
		maxRestarts:     maxRestarts,
		heartbeatTicker: time.NewTicker(10 * time.Second),
		watchdogTicker:  time.NewTicker(30 * time.Second),
		lastHeartbeat:   time.Now(),
	}
}

// SetRunning은 앱 실행 상태를 설정합니다
func (am *AppManager) SetRunning(running bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.isRunning = running
	if running {
		am.lastHeartbeat = time.Now()
	}
}

// IsRunning은 앱이 실행 중인지 확인합니다
func (am *AppManager) IsRunning() bool {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return am.isRunning
}

// UpdateHeartbeat는 앱의 마지막 하트비트 시간을 갱신합니다
func (am *AppManager) UpdateHeartbeat() {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.lastHeartbeat = time.Now()
}

// GetHeartbeatAge는 마지막 하트비트부터 현재까지의 경과 시간을 반환합니다
func (am *AppManager) GetHeartbeatAge() time.Duration {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return time.Since(am.lastHeartbeat)
}

// initializeApp은 애플리케이션을 초기화합니다
func (am *AppManager) initializeApp() error {
	// 컨테이너 생성
	container, err := container.NewContainer()
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	am.container = container

	// 컨텍스트 생성
	am.ctx, am.cancel = context.WithCancel(context.Background())

	// 애플리케이션 생성
	am.app = app.NewApplication(container, am.cancel)

	// 애플리케이션 초기화
	if err := am.app.Init(); err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	return nil
}

// startApp은 애플리케이션을 시작하고 패닉을 복구합니다
func (am *AppManager) startApp() {

	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Application panic occurred: %v\n%s", r, debug.Stack())
			am.SetRunning(false)
		}
	}()

	// 앱 초기화
	if err := am.initializeApp(); err != nil {
		logger.Printf("Failed to initialize application: %v", err)
		am.SetRunning(false)
		return
	}

	// 앱 시작
	if err := am.app.Start(); err != nil {
		logger.Printf("Failed to start application: %v", err)
		am.SetRunning(false)
		return
	}

	am.SetRunning(true)
	logger.Printf("Application started successfully (restart count: %d)", am.restartCount)
}

// startHeartbeat는 하트비트 고루틴을 시작합니다
func (am *AppManager) startHeartbeat() {
	am.wg.Add(1)
	go func() {
		defer am.wg.Done()
		defer am.heartbeatTicker.Stop()

		for {
			select {
			case <-am.heartbeatTicker.C:
				if am.IsRunning() {
					am.UpdateHeartbeat()
					logger.Printf("Application heartbeat updated")
				} else {
					// 앱이 실행 중이 아니면 하트비트 고루틴 종료
					return
				}
			case <-am.ctx.Done():
				// 컨텍스트가 취소되면 종료
				return
			}
		}
	}()
}

// startWatchdog는 감시견 고루틴을 시작합니다
func (am *AppManager) startWatchdog() {
	maxHeartbeatAge := 2 * time.Minute

	am.wg.Add(1)
	go func() {
		defer am.wg.Done()
		defer am.watchdogTicker.Stop()

		for {
			select {
			case <-am.watchdogTicker.C:
				// 앱이 실행 중이 아니거나 하트비트가 너무 오래된 경우
				if !am.IsRunning() || am.GetHeartbeatAge() > maxHeartbeatAge {
					if am.restartCount >= am.maxRestarts {
						logger.Printf("Maximum restart attempts (%d) reached. Stopping watchdog.", am.maxRestarts)
						am.cancel() // 전체 애플리케이션 종료
						return
					}

					if !am.IsRunning() {
						logger.Printf("Application is not running, attempting restart...")
					} else {
						logger.Printf("Application heartbeat too old (%v), attempting restart...", am.GetHeartbeatAge())
					}

					am.restartApp()
				}
			case <-am.ctx.Done():
				// 컨텍스트가 취소되면 종료
				return
			}
		}
	}()
}

// restartApp은 애플리케이션을 재시작합니다
func (am *AppManager) restartApp() {
	am.restartCount++
	logger.Printf("Restarting application (attempt %d/%d)", am.restartCount, am.maxRestarts)

	// 이전 앱 인스턴스 정리
	am.cleanupApp()

	// 잠시 대기 후 재시작
	time.Sleep(2 * time.Second)

	// 새 앱 시작
	go am.startApp()

	time.Sleep(1 * time.Second)

	// 새 하트비트 시작
	go am.startHeartbeat()
}

// cleanupApp은 현재 앱 인스턴스를 정리합니다
func (am *AppManager) cleanupApp() {
	if am.app != nil {
		logger.Printf("Shutting down previous application instance...")

		// 이전 컨텍스트 취소
		if am.cancel != nil {
			am.cancel()
		}

		am.app.Shutdown()

		// 최대 5초 동안 정상 종료 대기
		time.Sleep(5 * time.Second)
	}
}

// handleSignals는 시스템 시그널을 처리합니다
func (am *AppManager) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	am.wg.Add(1)
	go func() {
		defer am.wg.Done()

		<-sigChan
		logger.Printf("Shutdown signal received")
		am.Shutdown()
	}()
}

// Start는 애플리케이션 매니저를 시작합니다
func (am *AppManager) Start() {
	log.Println("Starting Application Manager...")

	// 시그널 핸들러 시작
	am.handleSignals()

	// 초기 앱 시작
	go am.startApp()

	time.Sleep(1 * time.Second)

	// 하트비트 시작
	go am.startHeartbeat()

	// 감시견 시작
	go am.startWatchdog()

	// 컨텍스트가 취소될 때까지 대기
	<-am.ctx.Done()
	logger.Printf("Application Manager context cancelled")
}

// Shutdown은 애플리케이션 매니저를 정상 종료합니다
func (am *AppManager) Shutdown() {
	logger.Printf("Initiating Application Manager shutdown...")

	// 실행 상태를 false로 설정
	am.SetRunning(false)

	// 컨텍스트 취소
	if am.cancel != nil {
		am.cancel()
	}

	// 현재 앱 인스턴스 정리
	am.cleanupApp()

	// 모든 고루틴이 종료될 때까지 대기
	am.wg.Wait()

	logger.Printf("Application Manager shutdown completed")
}

func main() {

	// 최상위 패닉 복구
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Unhandled panic in main: %v\n%s", r, debug.Stack())
			os.Exit(1)
		}
	}()

	// 애플리케이션 매니저 생성 (최대 5번 재시작 허용)
	appManager := NewAppManager(5)

	// 애플리케이션 매니저 시작
	appManager.Start()

	logger.Println("Application terminated")
}
