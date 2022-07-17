package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/ohhfeng/tinyRpc/codec"
	"github.com/ohhfeng/tinyRpc/codec/proto"
	"log"
	"net"
	"strings"
	"sync"
)

type Server struct {
	mu        sync.Mutex
	services  map[string]serviceInfo
	codecType string
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]serviceInfo),
	}
}

func (s *Server) Serve(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
		}
		go s.handleCoon(conn)
	}
}

func (s *Server) RegisterService(info ServiceInfo, serviceImpl interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[info.ServiceName]; ok {
		log.Fatalf("rpc:failed to register duplicate service,name:%s", info.ServiceName)
	}
	sinfo := serviceInfo{
		serviceImpl: serviceImpl,
		methods:     make(map[string]*Method),
	}
	for _, method := range info.Methods {
		sinfo.methods[method.Name] = method
	}
	s.services[info.ServiceName] = sinfo
}

func (s *Server) getCodec(name string) codec.Codec {
	if len(name) == 0 {
		return codec.Get(proto.Name)
	}
	return codec.Get(name)
}

func (s *Server) handleCoon(conn net.Conn) error {
	//todo 流错误处理，没想好怎么处理，可以返回Resp

	metaData, err := RevMetaData(conn)
	if err != nil {
		return err
	}
	mn := metaData.methodName
	if mn != "" && mn[0] == '/' {
		mn = mn[1:]
	}
	pos := strings.LastIndex(mn, "/")
	if pos == -1 {
		return errors.New(fmt.Sprintf("failed to parse method name,%s", mn))
	}
	serviceName := mn[:pos]
	methodName := mn[pos+1:]
	serviceInfo, ok := s.services[serviceName]
	if !ok {
		return errors.New(fmt.Sprintf("failed to find service name,%s", serviceName))
	}
	method, ok := serviceInfo.methods[methodName]
	if !ok {
		return errors.New(fmt.Sprintf("failed to find method name,%s", methodName))
	}
	// Unary
	err = s.processUnaryRPC(conn, method, serviceInfo.serviceImpl)
	if err != nil {
		return err
	}
	//Stream
	return nil
}

func (s *Server) processUnaryRPC(conn net.Conn, method *Method, srv interface{}) error {
	ctx := context.Background()
	c := s.getCodec(s.codecType)
	dec := func(m interface{}) error {
		err := RevcMsg(conn, c, m)
		if err != nil {
			return err
		}
		return nil
	}
	resp, err := method.Handler(ctx, srv, dec)
	if err != nil {
		return err
	}
	err = SendMsg(conn, resp, c)
	if err != nil {
		return err
	}
	return nil
}

type ServiceInfo struct {
	ServiceName string
	Methods     []*Method
}

type serviceInfo struct {
	serviceImpl interface{}
	methods     map[string]*Method
}

type methodHandler func(ctx context.Context, srv interface{}, dec func(v interface{}) error) (interface{}, error)

type Method struct {
	Name    string
	Handler methodHandler
}

type metadata struct {
	methodName string
}
