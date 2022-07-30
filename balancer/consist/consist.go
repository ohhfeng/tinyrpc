package consist

import (
	"errors"
	"github.com/ohhfeng/tinyRpc/balancer"
	"hash/crc32"
	"log"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

const (
	Name = "consist"
)

func init() {
	balancer.Register(&Builder{})
}

type Consist struct {
	replicas int64    //副本数
	hash     Hash     //哈希函数
	keys     []uint32 //记录包括虚拟节点的所有key
	hashmap  map[uint32]*balancer.ServiceInfo
}

func NewConsist(hash Hash, replicas int64) *Consist {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &Consist{
		replicas: replicas,
		hash:     hash,
		keys:     make([]uint32, 0),
		hashmap:  make(map[uint32]*balancer.ServiceInfo),
	}
}

type Builder struct{}

func (b *Builder) Builder() balancer.Balancer {
	return NewConsist(nil, 5)
}

func (b *Builder) Name() string {
	return Name
}

func (c *Consist) Add(infos ...*balancer.ServiceInfo) {
	for _, info := range infos {
		key := info.Addr()
		for i := int64(0); i < c.replicas; i++ {
			hashKey := c.hash([]byte(strconv.FormatInt(i, 10) + key))
			c.keys = append(c.keys, hashKey)
			c.hashmap[hashKey] = info
		}
	}
	sort.Slice(c.keys, func(i, j int) bool {
		return c.keys[i] < c.keys[j]
	})
}

func (c *Consist) Get(req *balancer.Request) (*balancer.ServiceInfo, error) {
	if len(c.keys) == 0 {
		return nil, errors.New("keys's length is 0")
	}
	key := req.String()
	hashKey := c.hash([]byte(key))
	idx := sort.Search(len(c.keys), func(i int) bool {
		return c.keys[i] >= hashKey
	})
	return c.hashmap[c.keys[idx%len(c.keys)]], nil
}

func (c *Consist) Remove(infos ...*balancer.ServiceInfo) {
	for _, info := range infos {
		key := info.Addr()
		for i := int64(0); i < c.replicas; i++ {
			hashKey := c.hash([]byte(strconv.FormatInt(i, 10) + key))
			idx := sort.Search(len(c.keys), func(i int) bool {
				return c.keys[i] >= hashKey
			})
			if idx == len(c.keys) || c.keys[idx] != hashKey {
				log.Printf("There is no such key:%s", key)
				return
			}
			c.keys = append(c.keys[:idx], c.keys[idx+1:]...)
			delete(c.hashmap, hashKey)
		}
	}
}
