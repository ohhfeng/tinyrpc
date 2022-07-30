package registry

import (
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/ohhfeng/tinyRpc/balancer"
	"log"
	"time"
)

const (
	ServicesRoot = "/services"
)

type Registry struct {
	service string
	info    *balancer.ServiceInfo
	stop    chan struct{}
	leaseId *clientv3.LeaseID
	client  *clientv3.Client
}

func NewRegistry(endpoints []string, service string, info *balancer.ServiceInfo) (*Registry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	r := &Registry{
		service: service,
		info:    info,
		stop:    make(chan struct{}),
		client:  cli,
	}
	err = r.register()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return r, nil
}

func (r *Registry) register() error {
	ch, err := r.keepAlive()
	if err != nil {
		log.Fatal(err)
		return err
	}
	go func() {
		for {
			select {
			case <-r.stop:
				r.revoke()
				log.Println("registry closed")
				return
			case <-r.client.Ctx().Done():
				log.Println("server closed")
				return
			case c, ok := <-ch:
				if !ok {
					log.Println("keep alive channel closed")
					r.revoke()
					return
				}
				log.Printf("service[%s,%s] revc reply,ttl:%d", r.service, r.info.Addr(), c.TTL)
			}
		}
	}()
	return nil
}

func (r *Registry) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	key := generateKey(r.service)
	value, _ := json.Marshal(r.info)

	resp, err := r.client.Grant(context.Background(), 5)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	r.leaseId = &resp.ID
	_, err = r.client.Put(context.Background(), key, string(value), clientv3.WithLease(*r.leaseId))
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return r.client.KeepAlive(context.Background(), *r.leaseId)
}

func (r *Registry) revoke() {
	_, err := r.client.Revoke(context.Background(), *r.leaseId)
	if err != nil {
		log.Printf("failed to revoke,err:%s", err.Error())
	} else {
		log.Printf("service:%s is stoped", r.service)
	}
}

func (r *Registry) Stop() {
	r.stop <- struct{}{}
}
