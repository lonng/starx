package protobuf

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

var ErrWrongValueType = errors.New("wrong value type")

type ProtobufSerialezer struct{}

func NewProtobufSerializer() *ProtobufSerialezer {
	return &ProtobufSerialezer{}
}

func (s *ProtobufSerialezer) Serialize(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {

	}
	return proto.Marshal(pb)
}

func (s *ProtobufSerialezer) Deserialize(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {

	}
	return proto.Unmarshal(data, pb)
}
