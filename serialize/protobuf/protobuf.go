package protobuf

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

var ErrWrongValueType = errors.New("struct must be converted to proto.Message")

type ProtobufSerialezer struct{}

func NewProtobufSerializer() *ProtobufSerialezer {
	return &ProtobufSerialezer{}
}

func (s *ProtobufSerialezer) Serialize(v interface{}) ([]byte, error) {
	pb, ok := v.(proto.Message)
	if !ok {
		return nil, ErrWrongValueType
	}
	return proto.Marshal(pb)
}

func (s *ProtobufSerialezer) Deserialize(data []byte, v interface{}) error {
	pb, ok := v.(proto.Message)
	if !ok {
		return ErrWrongValueType
	}
	return proto.Unmarshal(data, pb)
}
