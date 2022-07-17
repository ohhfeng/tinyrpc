package main

import (
	"context"
	"fmt"
	rpc "github.com/ohhfeng/tinyRpc"
	"github.com/ohhfeng/tinyRpc/helloworld/helloworld"
)

func main() {
	conn, err := rpc.Dail("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	request := helloworld.HelloRequest{Name: "111"}
	client := helloworld.NewHelloWorldClient(conn)
	reply, err := client.SayHello(context.Background(), &request)
	fmt.Println(reply, err)
}
