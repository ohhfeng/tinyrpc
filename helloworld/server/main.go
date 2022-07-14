package main

import (
	rpc "github.com/ohhfeng/tinyRpc"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	server := rpc.Server{}
	server.Serve(lis)
}
