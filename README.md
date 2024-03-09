# net package 

This Go package implements a simple, concurrent TCP chat server

## Build
```sh
go build main.go
```
### Sniffer
For sniffer you need to install pcap lib
```sh
sudo apt install libpcap-dev
```
## Usage
### Run TCP chat
```sh
main
```
### Run sniffer
you need to run sniffer with sudo.
```sh
sudo main sniffer
```

## Chat description
The server will listen on TCP port 6969 for incoming connections. 
Clients can connect to this port using any TCP client. 
For example, you can use telnet or netcat:

```sh
telnet localhost 6969
# or
nc localhost 6969
```

## Sniffer description
Sniffer listents to TCP port 6969 and prints to stdout information from IP (Network layer) and TCP (Transport layer)
 protocols about sender and receiver addreses and message payload.
