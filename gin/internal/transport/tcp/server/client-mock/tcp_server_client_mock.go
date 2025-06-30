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
	"time"
)

type RTMSMessage struct {
	STX      [2]byte
	Length   uint32
	Sequence byte
	UnitNo   byte
	OpCode   uint16
	Data     []byte
	LRC      byte
}

func (m *RTMSMessage) To0x10() []byte {
	b := make([]byte, 51)
	b[0], b[1] = m.STX[0], m.STX[1]
	binary.LittleEndian.PutUint32(b[2:6], m.Length)
	b[6], b[7] = m.Sequence, m.UnitNo
	binary.LittleEndian.PutUint16(b[8:10], m.OpCode)
	copy(b[10:50], m.Data)
	b[50] = m.LRC
	return b
}

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

// 이후에 테스트용으로 필요할지도
func (t *TCPClient) receiveMessage() {
	defer t.wg.Done()

	for {
		buf := make([]byte, 10)
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
			log.Println(string(buf))
		}
	}
}

func sendMessage(conn net.Conn, msg string) error {
	testMsg := &RTMSMessage{
		STX:      [2]byte{0x7E, 0x7E},
		Length:   51,
		Sequence: 1,
		UnitNo:   1,
		OpCode:   0,
		Data:     nil,
		LRC:      0,
	}

	if msg == "10" {
		testMsg.OpCode = 0x10
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
		testMsg.Data = data1
	}

	sendBytes, err := conn.Write(testMsg.To0x10())
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
