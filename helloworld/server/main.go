package main

import (
	"context"
	"fmt"
	rpc "github.com/ohhfeng/tinyRpc"
	"github.com/ohhfeng/tinyRpc/helloworld/helloworld"
	"net"
)

type hello struct {
}

func (h *hello) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	fmt.Println(req)
	return &helloworld.HelloReply{Message: "aaa"}, nil
}

func main() {
	//lis, err := net.Listen("tcp", "127.0.0.1:8080")
	//if err != nil {
	//	panic(err)
	//}
	//server := rpc.NewServer()
	//helloworld.RegisterHelloWorldServer(server, &hello{})
	//server.Serve(lis)
	lis, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	server := rpc.NewServer(rpc.WithRegistryEndpoints([]string{"127.0.0.1:2379"}))
	helloworld.RegisterHelloWorldServer(server, &hello{})
	server.Serve(lis)
}
