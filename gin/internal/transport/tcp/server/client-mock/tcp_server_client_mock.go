package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type TCPClient struct {
	conn        net.Conn
	sendMsgChan chan string
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
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (t *TCPClient) Start() {
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
			t.sendMsgChan <- line
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

func sendMessage(conn net.Conn, msg string) error {
	byteMsg := append([]byte(msg), '\n')
	_, err := conn.Write(byteMsg)
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
