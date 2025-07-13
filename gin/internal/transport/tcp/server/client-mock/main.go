package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"test/serializer.go"
	"time"
)

type TCPClient struct {
	conn        net.Conn
	sendMsgChan chan string
	readFull    chan struct{}
	mu          sync.RWMutex
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewTCPClient() (*TCPClient, error) {
	conn, err := net.DialTimeout("tcp", "localhost:5000", time.Second*5)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &TCPClient{
		conn:        conn,
		sendMsgChan: make(chan string),
		readFull:    make(chan struct{}, 1),
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (t *TCPClient) Start() {

	t.wg.Add(1)
	go t.receiveMessage()

	t.wg.Add(1)
	defer t.wg.Done()
	reader := bufio.NewReader(os.Stdin)
	for {
		go func() {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}
			t.sendMsgChan <- strings.TrimSpace(line)
		}()

		select {
		case <-t.ctx.Done():
			return
		case data := <-t.sendMsgChan:
			err := sendMessage(t.conn, data)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

/**
 * test: api -(eventbus)-> tcp server -(socket)-> client
 * api 제어 테스트
 */

func (t *TCPClient) receiveMessage() {
	defer t.wg.Done()

	for {
		buf := make([]byte, 1)
		go func() {
			_, err := io.ReadFull(t.conn, buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println(err)
			}
			t.readFull <- struct{}{}
		}()

		select {
		case <-t.ctx.Done():
			return
		case <-t.readFull:
			log.Printf("메시지 받음 OPCODE:0x%02X", buf[0])
			if buf[0] == 0x02 {
				sendMsg, err := serializer.SerializeRTMSResponse(binary.LittleEndian, 0x02, 1, nil)
				if err != nil {
					log.Println(err)
					continue
				}
				t.conn.Write(sendMsg)
			}
		}
	}
}

/**
 * test: client -> tcp server
 * 프로토콜 테스트
 */

func sendMessage(conn net.Conn, msg string) error {

	var testMsg []byte
	if msg == "10" {
		id := "chlwodns"
		passwd := "chlwodns123"
		data1 := make([]byte, 20)
		data2 := make([]byte, 20)
		for idx, ch := range id {
			data1[idx] = byte(ch)
		}
		for idx, ch := range passwd {
			data2[idx] = byte(ch)
		}
		data1 = append(data1, data2...)
		res, err := serializer.SerializeRTMSResponse(binary.LittleEndian, 0x10, 1, data1)
		if err != nil {
			log.Println(err)
			return err
		}
		testMsg = res
	}

	sendBytes, err := conn.Write(testMsg)
	log.Printf("보낸 바이트: %d", sendBytes)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (t *TCPClient) Shutdown() error {
	t.cancel()
	t.wg.Wait()
	return t.conn.Close()
}

func main() {
	tcpClient, err := NewTCPClient()
	if err != nil {
		log.Println(err)
		return
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go tcpClient.Start()

	<-signalChan
	log.Println("Shutting down...")
	tcpClient.Shutdown()
	log.Println("TCP Client terminated")
	close(signalChan)
}
