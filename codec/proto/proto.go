package proto

import (
	"fmt"
	"github.com/ohhfeng/tinyRpc/codec"
	"google.golang.org/protobuf/proto"
)

const Name = "proto"

type Codec struct{}

func init() {
	codec.Register(&Codec{})
}

func (p *Codec) Marshal(v interface{}) ([]byte, error) {
	message, ok := v.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T, want proto.Message", v)
	}
	return proto.Marshal(message)
}

func (p *Codec) Unmarshal(data []byte, v interface{}) error {
	message, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T, want proto.Message", v)
	}
	return proto.Unmarshal(data, message)
}

func (p *Codec) Name() string {
	return Name
}
