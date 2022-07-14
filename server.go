package rpc

import (
	"fmt"
	"github.com/ohhfeng/tinyRpc/codec"
	"github.com/ohhfeng/tinyRpc/codec/proto"
	"github.com/ohhfeng/tinyRpc/helloworld/helloworld"
	"log"
	"net"
)

type Server struct {
}

func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
		}
		go handleCoon(conn)
	}
}

func (s *Server) getCodec(name string) codec.Codec {
	if len(name) == 0 {
		return codec.Get(proto.Name)
	}
	return codec.Get(name)
}

func handleCoon(conn net.Conn) {
	//demo
	req := helloworld.HelloRequest{}
	err := RevcMsg(conn, codec.Get(proto.Name), &req)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(req.String())
	resp := helloworld.HelloReply{Message: "abc"}
	err = SendMsg(conn, &resp, codec.Get(proto.Name))
	if err != nil {
		fmt.Println(err)
	}
	// Unary

	//Stream
}
