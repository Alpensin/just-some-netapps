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

func finishConnection(conn net.Conn) {
	conn.Close()
	log.Printf("connection with %s was closed", conn.RemoteAddr())
}

func connReader(conn net.Conn, msgs chan string) {
	buffer := make([]byte, 8)
	for {
		n, err := conn.Read(buffer)
		log.Printf("read %d bytes for %s connection with %s", n, conn.RemoteAddr().Network(), conn.RemoteAddr())
		if err != nil {
			if err != io.EOF {
				log.Printf("could not read the input message")
				return
			}
		}
		msgs <- string(buffer[:n])
	}
}

func sendMessages(ac *activeConnections, msg []byte) {
	ac.rw.Lock()
	defer ac.rw.Unlock()
	for conn := range ac.conns {
		_, err := conn.Write(msg)
		if err != nil {
			log.Printf("could not write message to %s\n", conn.RemoteAddr())
		}
	}
}

// Будет не только публикация, но и менеджемент соединений. Информация о новых соединениях будет поступать по каналу.
func chatPublishing(ac *activeConnections, msgs <-chan string) {
	for {
		select {
		case msg := <-msgs:
			m := []byte(msg)
			sendMessages(ac, m)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func handleConnection(conn net.Conn, msgs chan string) {
	defer finishConnection(conn)
	log.Printf("new %s connection from address %s\n", conn.RemoteAddr().Network(), conn.RemoteAddr())
	message := []byte("Hello, friend!\n")
	n, err := conn.Write(message)
	if err != nil {
		log.Fatalf("could not write message\n")
		return
	}
	if n < len(message) {
		log.Printf("the message was not fully written to %s: expected:%d was written: %d\n",
			conn.RemoteAddr(), len(message), n)
		return
	}
	go connReader(conn, msgs)
	// echo server via channels
	for {
		select {
		case msg := <-msgs:
			m := []byte(msg)
			_, err := conn.Write(m)
			if err != nil {
				log.Fatalf("could not write message\n")
				return
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func funServer() {
	ln, err := net.Listen("tcp", ":"+tcpPort)
	if err != nil {
		log.Fatalf("could not listen to port %s: %s\n", tcpPort, err)
	}
	log.Printf("server started on port %s\n", tcpPort)
	msgs := make(chan string)
	go chatPublishing(&activeConnections{}, msgs)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("could not accept a connection: %s\n", err)
		}
		//  Добавить канал оповещения о новых соединениях
		go handleConnection(conn, msgs)
	}
}
