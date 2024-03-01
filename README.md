# net package 

This Go package implements a simple, concurrent TCP chat server

## Usage
```sh
go run main.go
```
The server will listen on TCP port 6969 for incoming connections. 
Clients can connect to this port using any TCP client. 
For example, you can use telnet or netcat:

```sh
telnet localhost 6969
# or
nc localhost 6969
```
