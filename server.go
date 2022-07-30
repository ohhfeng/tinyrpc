package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/ohhfeng/tinyRpc/balancer"
	"github.com/ohhfeng/tinyRpc/codec"
	"github.com/ohhfeng/tinyRpc/codec/proto"
	"github.com/ohhfeng/tinyRpc/registry"
	"log"
	"net"
	"strings"
	"sync"
)

type serverOptions struct {
	registry          bool
	registryEndpoints []string
}

type Option func(s *serverOptions)

func WithRegistryEndpoints(endpoints []string) Option {
	return func(s *serverOptions) {
		s.registry = true
		s.registryEndpoints = endpoints
	}
}

type Server struct {
	mu        sync.Mutex
	services  map[string]serviceInfo
	codecType string
	registry  []*registry.Registry
	opts      *serverOptions
}

func NewServer(ops ...Option) *Server {
	opts := &serverOptions{}
	for _, o := range ops {
		o(opts)
	}
	return &Server{
		opts:     opts,
		services: make(map[string]serviceInfo),
	}
}

func getHostAndPort(addr net.Addr) (host, port string, err error) {
	split := strings.Split(addr.String(), ":")
	if len(split) != 2 {
		return "", "", errors.New("the addr is illegal")
	}
	return split[0], split[1], nil
}

func (s *Server) register(addr net.Addr) {
	host, port, _ := getHostAndPort(addr)
	b := balancer.ServiceInfo{
		Host: host,
		Port: port,
	}
	if len(s.opts.registryEndpoints) != 0 {
		for k, _ := range s.services {
			r, err := registry.NewRegistry(s.opts.registryEndpoints, fmt.Sprintf("%s/%s:%s", k, host, port), &b)
			if err != nil {
				panic(err)
			}
			s.registry = append(s.registry, r)
		}
	}
}

func (s *Server) Serve(lis net.Listener) {
	if s.opts.registry {
		s.register(lis.Addr())
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			//todo 可参考http对临时错误做sleep处理
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
	//todo 流错误处理，没想好怎么处理，可以返回Resp,不应该抛出
	defer func() { _ = conn.Close() }()

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
