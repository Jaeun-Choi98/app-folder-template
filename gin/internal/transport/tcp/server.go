package tcp

import (
	"bufio"
	"context"
	"log"
	"net"
	"pjt/internal/logger"
	model "pjt/internal/model/event"
	"pjt/internal/transport/eventbus"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
}

type TCPServer struct {
	Listener net.Listener
	Clients  map[*Client]bool
	EventBus *eventbus.EventBus
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	wg       sync.WaitGroup
}

func NewTCPServer(eventBus *eventbus.EventBus) (*TCPServer, error) {
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		logger.Println(err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TCPServer{
		Listener: listener,
		Clients:  make(map[*Client]bool),
		EventBus: eventBus,
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

func (t *TCPServer) Start() error {
	// SendMessage goroutine을 닫기 위한 동기화
	// 클라이언트에게 다시 메시지를 보내야한다면 아래 코드 필요
	//wg.Add(1)
	//go SendMessage(ctx)

	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			log.Println(err)
			select {
			case <-t.ctx.Done():
				return err
			default:
			}
			continue
		}
		t.wg.Add(1)
		go t.HandleConnection(t.ctx, conn)
	}
}

func (t *TCPServer) HandleConnection(ctx context.Context, conn net.Conn) {

	client := &Client{conn: conn}
	defer func() {
		conn.Close()
		logger.Println("[TCP] Closed HandleConnection goruntine")
		t.wg.Done()
	}()
	t.mu.Lock()
	t.Clients[client] = true
	t.mu.Unlock()

	handleMessage := func(msg string) {
		switch msg {
		case "EVENTA":
			t.EventBus.Publish(model.EVENTA, &model.EventA{Type: model.EVENTA, Payload: model.MessageA{Name: "cju", Age: 11}})
			logger.Println("[TCP] AAAAAAAAAAAAA")
		case "EVENTB":
			t.EventBus.Publish(model.EVENTB, &model.EventB{Type: model.EVENTB, Payload: model.MessageB{Args: []string{"sdf", "sdf"}, Cmd: "stres"}})
			logger.Println("[TCP] BBBBBBBBBBBBB")
		default:
			logger.Println("[TCP] Unknown Message Type")
		}
	}

	receiveMessage := func() {
		reader := bufio.NewReader(conn)
		receiveMsgChan := make(chan string, 1)
		defer func() {
			close(receiveMsgChan)
			logger.Println("[TCP] Closed receiveMessage func")
		}()

		closedConnByRemoteHost := make(chan bool, 1)
		defer close(closedConnByRemoteHost)

		for {
			// if an existing connection was forcibly closed by the remote host, break select block
			go func() {
				line, err := reader.ReadString('\n')

				if err != nil {
					logger.Println(err)
					closedConnByRemoteHost <- true
					return
				}
				msg := strings.TrimSpace(line)
				receiveMsgChan <- msg
			}()

			select {
			case <-ctx.Done():
				return
			case <-closedConnByRemoteHost:
				return
			case msg := <-receiveMsgChan:
				handleMessage(msg)
			}
		}
	}

	logger.Printf("[TCP] Connected new client: %s\n", conn.RemoteAddr().String())
	receiveMessage()
	t.mu.Lock()
	delete(t.Clients, client)
	t.mu.Unlock()
	logger.Printf("[TCP] Disconnected client: %s\n", conn.RemoteAddr().String())
}

func (t *TCPServer) Shutdown() error {
	t.cancel()
	t.wg.Wait()
	logger.Println("[TCP] TCP goroutine terminated")
	return t.Listener.Close()
}

// func SendMessage(ctx context.Context) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			// 정상 종료 시에 SendMessage goroutine이 제대로 닫히는지 확인.
// 			log.Println("closed SendMessage goruntine")
// 			wg.Done()
// 			return
// 			// 여기에 이벤트 버스 구독 채널이 와야함
// 		case strMsg := <-messageChan:
// 			mu.RLock()
// 			for client, exists := range clients {
// 				if exists {
// 					client.conn.Write([]byte(strMsg))
// 				}
// 			}
// 			mu.RUnlock()
// 		}
// 	}
// }
