package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/ohhfeng/tinyRpc/balancer"
	"log"
	"sync"
	"time"
)

type Discovery struct {
	client   *clientv3.Client
	mu       sync.Mutex
	balancer balancer.Balancer
}

func NewDiscovery(endpoints []string, serviceName, balancerName string) (*Discovery, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	builder := balancer.Get(balancerName)
	d := &Discovery{
		client:   cli,
		balancer: builder.Builder(),
	}
	//ch 用于当获取完服务列表之后，再消费watch chan中的数据。
	//防止先获取服务列表,再去调用watch方法消费watch chan,这两步之间出现服务注册或注销的消息缺失。
	ch := make(chan struct{})
	go d.watchServiceUpdate(ch)
	err = d.getServices(ch, serviceName)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return d, nil
}

func generateKey(serviceName string) string {
	return fmt.Sprintf("%s/%s", ServicesRoot, serviceName)
}

func (d *Discovery) getServices(ch chan struct{}, serviceName string) error {
	resp, err := d.client.Get(context.Background(), generateKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, kv := range resp.Kvs {
		info, err := unmarshalInfo(kv.Value)
		if err != nil {
			log.Fatal(err)
			return err
		}
		d.balancer.Add(info)
	}
	ch <- struct{}{}
	return nil
}

func (d *Discovery) watchServiceUpdate(ch chan struct{}) {
	watch := d.client.Watch(context.Background(), ServicesRoot, clientv3.WithPrefix(), clientv3.WithPrevKV())
	<-ch
	for resp := range watch {
		for _, ev := range resp.Events {
			switch ev.Type {
			case clientv3.EventTypePut:
				fmt.Printf("[%s] %q :%q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
				info, err := unmarshalInfo(ev.Kv.Value)
				if err != nil {
					continue
				}
				d.balancer.Add(info)
			case clientv3.EventTypeDelete:
				fmt.Printf("[%s] %q :%q\n", ev.Type, ev.PrevKv.Key, ev.PrevKv.Value)
				info, err := unmarshalInfo(ev.PrevKv.Value)
				if err != nil {
					continue
				}
				d.balancer.Remove(info)
			}
		}
	}
}

func unmarshalInfo(bytes []byte) (*balancer.ServiceInfo, error) {
	info := &balancer.ServiceInfo{}
	err := json.Unmarshal(bytes, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (d *Discovery) Get(request *balancer.Request) (*balancer.ServiceInfo, error) {
	return d.balancer.Get(request)
}
