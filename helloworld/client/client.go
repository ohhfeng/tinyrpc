package main

import (
	"fmt"
	rpc "github.com/ohhfeng/tinyRpc"
	"github.com/ohhfeng/tinyRpc/codec"
	"github.com/ohhfeng/tinyRpc/codec/proto"
	"github.com/ohhfeng/tinyRpc/helloworld/helloworld"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	request := helloworld.HelloRequest{Name: "111"}
	err = rpc.SendMsg(conn, &request, codec.Get(proto.Name))
	if err != nil {
		fmt.Println(err)
	}

	resp := helloworld.HelloReply{}
	err = rpc.RevcMsg(conn, codec.Get(proto.Name), &resp)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.String())
}
