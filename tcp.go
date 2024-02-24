// tcp server fun
package main

import (
	"io"
	"log"
	"net"
	"time"
)

const tcpPort = "6969"

func finishConnection(conn net.Conn) {
	conn.Close()
	log.Printf("Connection with %s was closed", conn.RemoteAddr())
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("could not accept a connection: %s\n", err)
		}
		go handleConnection(conn, msgs)
	}
}
