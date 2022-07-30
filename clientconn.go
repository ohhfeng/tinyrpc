package rpc

import (
	"context"
	"github.com/google/uuid"
	"github.com/ohhfeng/tinyRpc/balancer"
	"github.com/ohhfeng/tinyRpc/balancer/consist"
	"github.com/ohhfeng/tinyRpc/codec"
	"github.com/ohhfeng/tinyRpc/codec/proto"
	"github.com/ohhfeng/tinyRpc/registry"
	"net"
)

type Conn struct {
	conn      net.Conn
	codecType string
	id        string
}

func (c *Conn) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}
func (c *Conn) getCodec(name string) codec.Codec {
	if len(name) == 0 {
		return codec.Get(proto.Name)
	}
	return codec.Get(name)
}

func (c *Conn) Invoke(ctx context.Context, req interface{}, resp interface{}, methodName string) error {
	m := metadata{methodName: methodName}
	err := SendMetaData(c, m)
	if err != nil {
		return err
	}
	cc := c.getCodec(c.codecType)
	err = SendMsg(c, req, cc)
	if err != nil {
		return err
	}
	err = RevcMsg(c, cc, resp)
	if err != nil {
		return err
	}
	return nil
}

func Dail(address string) (*Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &Conn{conn: conn}, err
}

func XDail(serviceName string, endpoint []string) (*Conn, error) {
	discovery, err := registry.NewDiscovery(endpoint, serviceName, consist.Name)
	if err != nil {
		return nil, err
	}
	connId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	info, err := discovery.Get(&balancer.Request{ClientId: connId.String()})
	if err != nil {
		return nil, err
	}
	return Dail(info.Addr())
}
