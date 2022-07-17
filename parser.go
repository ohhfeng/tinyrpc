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

// +---------------+-------------+-------------+-------------+-------------+
// |    metadata   |   request   |    msg1     |     msg2    |   response  |
// +---------------+-------------+-------------+-------------+-------------+
// |     method    | head | body | head | body | head | body | head | body |
// +---------------+-------------+-------------+-------------+-------------+
// | length | name |length|      |length|      |length|      |length|      |
// +---------------+-------------+-------------+-------------+-------------+
// response可以在msg之前

func SendMetaData(w io.Writer, m metadata) error {
	header := prepareLen([]byte(m.methodName))
	_, err := w.Write(header)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(m.methodName))
	if err != nil {
		return err
	}
	return nil
}

func RevMetaData(r io.Reader) (*metadata, error) {
	hrd := make([]byte, sizeLen)
	_, err := r.Read(hrd)
	if err != nil {
		log.Println("failed to read metadata header")
		return nil, err
	}

	length := binary.BigEndian.Uint64(hrd)
	data := make([]byte, int(length))
	if _, err = r.Read(data); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	m := &metadata{methodName: string(data)}
	return m, nil
}

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
	header = prepareLen(body)
	return header, body, nil
}

func prepareLen(data []byte) (header []byte) {
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
