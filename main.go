// some net fun
package main

import (
	"fmt"
	"os"

	"github.com/alpensin/netfun/chat"
	"github.com/alpensin/netfun/sniffer"
)

func main() {
	args := os.Args
	if len(args) == 1 || args[1] == "chat" {
		chat.TCPChatServer()
	}
	if args[1] == "sniffer" {
		sniffer.Sniff()
	}
	panic(fmt.Sprintf("unexpeced arguments: %v", args))
}
