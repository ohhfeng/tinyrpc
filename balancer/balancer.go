package balancer

import (
	"fmt"
)

type Builder interface {
	Builder() Balancer
	Name() string
}

type Balancer interface {
	Add(infos ...*ServiceInfo)
	Get(request *Request) (*ServiceInfo, error)
	Remove(infos ...*ServiceInfo)
}
type Map map[string]Builder

func (m Map) register(c Builder) {
	m[c.Name()] = c
}

func (m Map) get(name string) Builder {
	return m[name]
}

func Register(c Builder) {
	balancerMap.register(c)
}

func Get(name string) Builder {
	return balancerMap.get(name)
}

var balancerMap Map = make(map[string]Builder, 0)

type ServiceInfo struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	Weight int64  `json:"weight"`
}

func (n *ServiceInfo) Addr() string {
	return fmt.Sprintf("%s:%s", n.Host, n.Port)
}

type Request struct {
	ClientId string
}

func (r *Request) String() string {
	return fmt.Sprintf("%s", r.ClientId)
}
