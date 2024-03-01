// Package chat server fun
package chat

import (
	"io"
	"log"
	"net"
	"sync"
)

const (
	tcpPort    = "6969"
	bufferSize = 256
)

type activeConnections struct {
	conns map[net.Conn]struct{}
	rw    sync.RWMutex
}

type connMessage struct {
	text []byte
	conn net.Conn
}

func newActiveConnections() *activeConnections {
	return &activeConnections{
		conns: map[net.Conn]struct{}{},
		rw:    sync.RWMutex{},
	}
}

func finishConnection(conn net.Conn, closedConnectionsCh chan<- net.Conn) {
	closedConnectionsCh <- conn
	conn.Close()
	log.Printf("connection with %s was closed", conn.RemoteAddr())
}

func connReader(conn net.Conn, msgsCh chan<- connMessage) {
	buffer := make([]byte, bufferSize)
	for {
		n, err := conn.Read(buffer)
		log.Printf("read %d bytes for %s connection with %s", n, conn.RemoteAddr().Network(), conn.RemoteAddr())
		if err != nil {
			if err != io.EOF {
				log.Printf("could not read the input message - %s", err)
				return
			}
			log.Print("received EOF", err)
			return
		}
		msgsCh <- connMessage{
			text: buffer[:n],
			conn: conn,
		}
	}
}

func sendMessages(ac *activeConnections, msg connMessage) {
	ac.rw.Lock()
	defer ac.rw.Unlock()
	for conn := range ac.conns {
		if conn == msg.conn {
			continue
		}
		_, err := conn.Write(msg.text)
		if err != nil {
			log.Printf("could not write message to %s", conn.RemoteAddr())
		}
	}
}

func chatManaging(ac *activeConnections, msgsCh <-chan connMessage, newConnectionsCh <-chan net.Conn, closedConnectionsCh <-chan net.Conn) {
	for {
		select {
		case msg := <-msgsCh:
			sendMessages(ac, msg)
		case newConnection := <-newConnectionsCh:
			log.Printf("new connection: %s", newConnection.RemoteAddr())
			ac.rw.Lock()
			ac.conns[newConnection] = struct{}{}
			ac.rw.Unlock()
		case closedConnection := <-closedConnectionsCh:
			log.Printf("closed connection: %s", closedConnection.RemoteAddr())
			ac.rw.Lock()
			log.Printf("closedConnection to delete: %s", closedConnection.RemoteAddr())
			delete(ac.conns, closedConnection)
			log.Printf("Active connections left: %d", len(ac.conns))
			ac.rw.Unlock()
		}
	}
}

func handleConnection(conn net.Conn, msgsCh chan<- connMessage, closedConnectionsCh chan<- net.Conn) {
	defer finishConnection(conn, closedConnectionsCh)
	log.Printf("new %s connection from address %s", conn.RemoteAddr().Network(), conn.RemoteAddr())
	message := []byte("Hello, friend!\n")
	n, err := conn.Write(message)
	if err != nil {
		log.Fatalf("could not write message")
		return
	}
	if n < len(message) {
		log.Printf("the message was not fully written to %s: expected:%d was written: %d",
			conn.RemoteAddr(), len(message), n)
		return
	}
	connReader(conn, msgsCh)
}

// TCPChatServer - simple TCP chat server
func TCPChatServer() {
	ln, err := net.Listen("tcp", ":"+tcpPort)
	defer func() {
		err := ln.Close()
		if err != nil {
			log.Fatal("listener faild to close")
		}
	}()

	if err != nil {
		log.Fatalf("could not listen to port %s: %s", tcpPort, err)
	}
	log.Printf("server started on port %s", tcpPort)
	msgsCh := make(chan connMessage)
	newConnectionsCh := make(chan net.Conn)
	closedConnectionsCh := make(chan net.Conn)
	go chatManaging(newActiveConnections(), msgsCh, newConnectionsCh, closedConnectionsCh)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("could not accept a connection: %s", err)
			continue
		}
		newConnectionsCh <- conn
		go handleConnection(conn, msgsCh, closedConnectionsCh)
	}
}
