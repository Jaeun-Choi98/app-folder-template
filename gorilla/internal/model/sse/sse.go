package model

/**
 * Event는 클라이언트에게 전송될 이벤트를 정의합니다.
 * SSE 프로토콜에 맞게 구조화되어 있습니다.
 */
type SSEMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
