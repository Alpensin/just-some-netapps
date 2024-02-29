// tcp server fun
package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const tcpPort = "6969"

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
	buffer := make([]byte, 16)
	for {
		n, err := conn.Read(buffer)
		log.Printf("read %d bytes for %s connection with %s", n, conn.RemoteAddr().Network(), conn.RemoteAddr())
		if err != nil {
			if err != io.EOF {
				log.Printf("could not read the input message - %s", err)
				return
			} else {
				log.Print("received EOF", err)
				return
			}
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
			log.Printf("new message: %s", msg)
			sendMessages(ac, msg)
		case newConnection := <-newConnectionsCh:
			log.Printf("new connection: %s", newConnection)
			ac.rw.Lock()
			ac.conns[newConnection] = struct{}{}
			ac.rw.Unlock()
		case closedConnection := <-closedConnectionsCh:
			log.Printf("closed connection: %s", closedConnection)
			ac.rw.Lock()
			log.Printf("ac.conns before delete: %#v", ac.conns)
			log.Printf("closedConnection to delete: %#v", closedConnection)
			delete(ac.conns, closedConnection)
			log.Printf("ac.conns after delete: %#v", ac.conns)
			ac.rw.Unlock()
		default:
			// log.Printf("connections: %#v", ac.conns)
			time.Sleep(1 * time.Millisecond)
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

func funServer() {
	ln, err := net.Listen("tcp", ":"+tcpPort)
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
		newConnectionsCh <- conn
		if err != nil {
			log.Printf("could not accept a connection: %s", err)
		}
		go handleConnection(conn, msgsCh, closedConnectionsCh)
	}
}
