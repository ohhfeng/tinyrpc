package helloworld

//todo 为了调试，暂时手写，后续改为插件生成
import (
	"context"
	rpc "github.com/ohhfeng/tinyRpc"
)

type HelloWorldClient struct {
	conn *rpc.Conn
}

func NewHelloWorldClient(conn *rpc.Conn) HelloWorldClient {
	return HelloWorldClient{conn}
}

func (h *HelloWorldClient) SayHello(ctx context.Context, request *HelloRequest) (*HelloReply, error) {
	response := &HelloReply{}
	err := h.conn.Invoke(ctx, request, response, "/HelloWorld/SayHello")
	if err != nil {
		return nil, err
	}
	return response, nil
}

type HelloWorldServerImpl interface {
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
}

func RegisterHelloWorldServer(server *rpc.Server, impl HelloWorldServerImpl) {
	server.RegisterService(serviceInfo, impl)
}

func HelloWorldServerHanlder(ctx context.Context, srv interface{}, dec func(v interface{}) error) (interface{}, error) {
	req := &HelloRequest{}
	err := dec(req)
	if err != nil {
		return nil, err
	}
	return srv.(HelloWorldServerImpl).SayHello(ctx, req)
}

var serviceInfo = rpc.ServiceInfo{
	ServiceName: "HelloWorld",
	Methods: []*rpc.Method{
		{
			Name:    "SayHello",
			Handler: HelloWorldServerHanlder,
		},
	},
}
