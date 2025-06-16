package sse

import (
	"context"
	"fmt"
	"net/http"
	"pjt/internal/logger"
)

/**
 * SSEClient는 SSE 연결을 유지하는 개별 클라이언트를 나타냄
 * 각 클라이언트는 고유 ID와 응답 Writer를 가짐
 */
type SSEClient struct {
	ClientId string
	UserId   string
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Ctx      context.Context
	Cancel   context.CancelFunc
}

// SSEService의 메서드에서 구현.
// func NewSSEClient(clientId, userId string, ctx *gin.Context) (*SSEClient, error) {
// 	flusher, ok := ctx.Writer.(http.Flusher)
// 	if !ok {
// 		return nil, fmt.Errorf("streaming not supported")
// 	}
// 	clientCtx, cancel := context.WithCancel(ctx.Request.Context())
// 	return &SSEClient{
// 		ClientId: clientId,
// 		UserId:   userId,
// 		Writer:   ctx.Writer,
// 		Flusher:  flusher,
// 		Ctx:      clientCtx,
// 		Cancel:   cancel,
// 	}, nil
// }

// SendEvent는 SSE 클라이언트에게 이벤트를 전송
func (c *SSEClient) SendMessage(msg Message) error {
	// 컨텍스트가 취소되었는지 확인
	select {
	case <-c.Ctx.Done():
		err := fmt.Errorf("client connection closed")
		logger.Printf("Failed to send message:\n\t%v", err)
		return err
	default:
		// SSE 형식으로 이벤트 전송
		fmt.Fprintf(c.Writer, "event: %s\n", msg.Type)
		fmt.Fprintf(c.Writer, "data: %v\n\n", msg.Payload)
		c.Flusher.Flush()
		return nil
	}
}

// Close는 클라이언트 연결을 종료
func (c *SSEClient) Close() {
	c.Cancel()
}
