package rpc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/ohhfeng/tinyRpc/codec"
	"io"
	"log"
	"math"
)

const (
	sizeLen = 8
)

func SendMsg(w io.Writer, m interface{}, c codec.Codec) error {
	header, data, err := prepareMsg(m, c)
	if err != nil {
		return err
	}
	_, err = w.Write(header)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func prepareMsg(m interface{}, c codec.Codec) (header []byte, data []byte, err error) {
	body, err := c.Marshal(m)
	if err != nil {
		return nil, nil, err
	}
	if uint(len(body)) > math.MaxUint64 {
		return nil, nil, errors.New(fmt.Sprintf("message too large (%d bytes)", len(body)))
	}
	header = prepareHeader(body)
	return header, body, nil
}

func prepareHeader(data []byte) (header []byte) {
	hdr := make([]byte, sizeLen)
	binary.BigEndian.PutUint64(hdr, uint64(len(data)))
	return hdr
}

func RevcMsg(r io.Reader, c codec.Codec, m interface{}) error {
	hrd := make([]byte, sizeLen)
	_, err := r.Read(hrd)
	if err != nil {
		log.Println("failed to read header")
		return err
	}

	length := binary.BigEndian.Uint64(hrd)
	data := make([]byte, int(length))
	if _, err = r.Read(data); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return err
	}

	err = c.Unmarshal(data, m)
	if err != nil {
		return err
	}
	return nil
}
