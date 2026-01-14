package service

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"pjt/internal/infra/cache"
	"pjt/internal/logger"

	"sync"
	"time"
)

/**
 * 파일 전송 과정에서 앱 단위의 패킷 처리를 위한 session
 *
 */

type TransferState int

const (
	StateIdle         TransferState = iota // 생성됨, 시작 전
	StateWaitStartAck                      // 시작 ACK 대기 중 (5초)
	StateSending                           // 윈도우 전송 중
	StateWaitCheckAck                      // ACK 대기 중 (5초)
	StateWaiteEndAck                       // 종료 ACK 대기 중 (5초)
	StateCompleted                         // 전송 완료
	StateTimeout                           // ACK 타임아웃
	StateCancelled                         // 취소됨
	StateFailed                            // 메시지 전송 실패
)

// // 에러 정의
var (
	ErrTimeout        = fmt.Errorf("ack timeout")
	ErrCancelled      = fmt.Errorf("transfer cancelled")
	ErrSessionClosed  = fmt.Errorf("session closed")
	ErrTransmitFailed = fmt.Errorf("transmit failed or panic")
)

type FileTransferJob struct {
	// 식별 정보
	terminalId uint32
	filepath   string

	// 파일 정보
	totalChunks int
	fileSize    int64

	// 전송 설정
	chunkSize  int
	windowSize int
	ackTimeout time.Duration

	// 전송 상태
	state          TransferState
	lastAckedChunk int // 클라이언트가 ACK한 마지막 청크 인덱스

	// 동시성 제어
	mu sync.Mutex

	// 0x2B를 받기 위한 채널
	startChan chan struct{}
	// 0x2C를 받기 위한 채널
	checkChan chan int
	// 0x2D를 받기 위한 채널
	endChan chan struct{}

	// 의존성
	cache *cache.Cache
}

func NewFileTransferSession(
	terminalId uint32,
	filepath string,
	cache *cache.Cache,
	chunkSize int,
	windowSize int,
) (*FileTransferJob, error) {

	var fileSize int64
	cache.MuFileInfoCache.RLock()
	if _, exists := cache.FileInfoCache[filepath]; exists {
		fileSize = cache.FileInfoCache[filepath].FileSize
	}
	cache.MuFileInfoCache.RUnlock()

	if fileSize == 0 || chunkSize == 0 || windowSize == 0 {
		return nil, fmt.Errorf("[File Transfer] wrong params, file size: %v cunk size: %v, window size: %v",
			fileSize, chunkSize, windowSize)
	}

	totalChunks := int(math.Ceil(float64(fileSize) / float64(chunkSize)))

	return &FileTransferJob{
		terminalId:     terminalId,
		filepath:       filepath,
		totalChunks:    totalChunks,
		fileSize:       fileSize,
		chunkSize:      chunkSize,
		windowSize:     windowSize,
		ackTimeout:     5 * time.Second,
		state:          StateIdle,
		lastAckedChunk: -1,
		startChan:      make(chan struct{}, 1),
		checkChan:      make(chan int, 10),
		endChan:        make(chan struct{}, 1),
		cache:          cache,
	}, nil
}

func (s *FileTransferJob) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateCancelled
	close(s.startChan)
	close(s.checkChan)
	close(s.endChan)
	return nil
}

func (s *FileTransferJob) GetState() TransferState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state
}

func (s *FileTransferJob) SetState(state TransferState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf(
		"[debug, File Transfer] file transfer state: %v(0: idle, 1:wait start ack, 2: sending, 3: check ack, 4: end ack, 5: completed, 6: timeout, 7: cancelled, 8: failed), terminalid: %v",
		state, s.terminalId)
	s.state = state
}

// 현재 파일을 전송하고 있는지 확인.
func (s *FileTransferJob) IsDone() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.state == StateIdle ||
		s.state == StateCompleted ||
		s.state == StateTimeout ||
		s.state == StateCancelled ||
		s.state == StateFailed
}

// 마지막 청크까지 ACK 받았는지 확인
func (s *FileTransferJob) isComplete() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.lastAckedChunk >= s.totalChunks-1
}

func (s *FileTransferJob) StartAck() error {
	s.mu.Lock()
	currentState := s.state
	s.mu.Unlock()

	// StateWaitStartAck 상태가 아니면 무시
	if currentState != StateWaitStartAck {
		s.SetState(StateFailed)
		return fmt.Errorf("received start ACK in state %v, ignored, terminal: %v", currentState, s.terminalId)
	}

	select {
	case s.startChan <- struct{}{}:
		return nil
	default:
		s.SetState(StateFailed)
		return fmt.Errorf("start ack channel full")
	}
}

func (s *FileTransferJob) CheckAck(ackIdx int) error {
	s.mu.Lock()
	currentState := s.state
	s.mu.Unlock()

	// StateWaitCheckAck 상태가 아니면 무시
	if currentState != StateWaitCheckAck {
		s.SetState(StateFailed)
		return fmt.Errorf("received file transfer check ACK in state %v, ignored, terminal: %v", currentState, s.terminalId)
	}

	select {
	case s.checkChan <- ackIdx:
		return nil
	default:
		s.SetState(StateFailed)
		return fmt.Errorf("check ack channel full")
	}
}

func (s *FileTransferJob) EndACK() error {
	s.mu.Lock()
	currentState := s.state
	s.mu.Unlock()

	// StateWaitEndAck 상태가 아니면 무시
	if currentState != StateWaiteEndAck {
		s.SetState(StateFailed)
		return fmt.Errorf("received end ACK in state %v, ignored, terminal: %v", currentState, s.terminalId)
	}

	select {
	case s.endChan <- struct{}{}:
		return nil
	default:
		s.SetState(StateFailed)
		return fmt.Errorf("end ack channel full, terminal: %v", s.terminalId)
	}
}

// Run 메서드 예시
func (s *FileTransferJob) Run(sendFunc func(clientId uint32, opCode byte, buf []byte, sendTimeout time.Duration) error) (retErr error) {

	defer func() {
		if r := recover(); r != nil {
			logger.Printf("[File Transfer] panic in running file transfer session, terminal: %v, panic: %v", s.terminalId, r)
			retErr = fmt.Errorf("[File Transfer] panic in running file transfer session, panic: %v", r)
			s.SetState(StateFailed)
		}
	}()

	s.lastAckedChunk = -1
	s.SetState(StateWaitStartAck)

	// 0x2B 대기 (5초)
	select {
	case <-s.startChan:
		// log.Printf("[debug, File Transfer] start ack, terminalid: %v", s.terminalId)
		// 정상
	case <-time.After(s.ackTimeout):

		if s.GetState() == StateCancelled {
			retErr = ErrCancelled
			return
		}

		s.SetState(StateTimeout)
		// 0x1D 전송 - 응답없음
		data := make([]byte, 5)
		binary.BigEndian.PutUint32(data[0:4], s.terminalId)
		data[4] = 1
		if err := sendFunc(s.terminalId, 0x1D, data, 0); err != nil {
			//logger.Printf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
			retErr = fmt.Errorf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
			s.SetState(StateFailed)
			return
		}
		retErr = fmt.Errorf("[File Transfer] start ack timeout, terminal: %v", s.terminalId)
		return
	}

	if s.GetState() == StateCancelled {
		retErr = ErrCancelled
		return
	}

	s.SetState(StateSending)

	for {
		if s.GetState() == StateCancelled {
			retErr = ErrCancelled
			return
		}

		// 윈도우 전송
		if err := s.sendWindow(sendFunc); err != nil {
			s.SetState(StateFailed)
			return err
		}

		s.SetState(StateWaitCheckAck)

		// 0x2C 대기 (5초)
		select {
		case ackIdx := <-s.checkChan:

			if s.GetState() == StateCancelled {
				retErr = ErrCancelled
				return
			}

			s.handleAckIndex(ackIdx)

			if s.isComplete() {
				// 0x1D 전송 - 정상
				data := make([]byte, 5)
				binary.BigEndian.PutUint32(data[0:4], s.terminalId)
				data[4] = 1
				if err := sendFunc(s.terminalId, 0x1D, data, 0); err != nil {
					//	logger.Printf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
					retErr = fmt.Errorf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
					s.SetState(StateFailed)
					return
				}

				// 0x2D 대기
				s.SetState(StateWaiteEndAck)

				// 파일 전송이 끝남.
				select {
				case <-s.endChan:
					s.SetState(StateCompleted)
					return nil
				case <-time.After(s.ackTimeout):
					s.SetState(StateTimeout)
					retErr = fmt.Errorf("[File Transfer] end ack timeout, terminal: %v", s.terminalId)
					return nil
				}
			}

			s.SetState(StateSending)

		case <-time.After(s.ackTimeout):

			if s.GetState() == StateCancelled {
				retErr = ErrCancelled
				return
			}

			s.SetState(StateTimeout)
			// 0x1D 전송 - 응답없음
			data := make([]byte, 5)
			binary.BigEndian.PutUint32(data[0:4], s.terminalId)
			data[4] = 1
			if err := sendFunc(s.terminalId, 0x1D, data, 0); err != nil {
				//logger.Printf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
				retErr = fmt.Errorf("[File Transfer] Failed to send 0x1D to terminal: %v", s.terminalId)
				s.SetState(StateFailed)
				return
			}
			return fmt.Errorf("[File Transfer] check ack timeout, terminal: %v", s.terminalId)
		}
	}
}

func (s *FileTransferJob) sendWindow(sendFunc func(clientId uint32, opCode byte, buf []byte, sendTimeout time.Duration) error) error {
	s.mu.Lock()
	startIdx := s.lastAckedChunk + 1
	endIdx := startIdx + s.windowSize

	if endIdx > s.totalChunks {
		endIdx = s.totalChunks
	}
	s.mu.Unlock()

	log.Printf("[debug, File Transfer] sending window [%d-%d] to terminal: %v", startIdx, endIdx-1, s.terminalId)

	for chunkIdx := startIdx; chunkIdx < endIdx; chunkIdx++ {
		chunk, err := s.cache.ReadChunk(s.filepath, chunkIdx, s.chunkSize)
		if err != nil {
			return fmt.Errorf("read chunk %d failed: %w", chunkIdx, err)
		}
		// 0x1B: 파일전송
		data := make([]byte, 12)
		binary.BigEndian.PutUint32(data[0:4], s.terminalId)
		binary.BigEndian.PutUint32(data[4:8], uint32(chunkIdx))
		binary.BigEndian.PutUint32(data[8:12], uint32(len(chunk)))
		data = append(data, chunk...)
		if err := sendFunc(s.terminalId, 0x1B, data, 0); err != nil {
			//logger.Printf("[File Transfer] Failed to send 0x1B to terminal: %v", s.terminalId)
			return fmt.Errorf("[File Transfer] Failed to send 0x1B to terminal: %v", s.terminalId)
		}
	}
	return nil
}

// ACK 인덱스 처리
func (s *FileTransferJob) handleAckIndex(ackIdx int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("[debug, File Transfer] received ACK for chunk %d from terminal %v", ackIdx, s.terminalId)

	s.lastAckedChunk = ackIdx
}
