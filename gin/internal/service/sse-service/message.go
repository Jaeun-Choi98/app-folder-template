package sse

/**
 * Event는 클라이언트에게 전송될 이벤트를 정의합니다.
 */
type Message struct {
	Type    string
	Payload interface{}
}
