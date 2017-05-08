package protobuf

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
)

type Message struct {
	Data *string `protobuf:"bytes,1,name=data"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}

func TestProtobufSerialezer_Serialize(t *testing.T) {
	m := &Message{proto.String("hello")}
	s := NewSerializer()

	b, err := s.Serialize(m)
	if err != nil {
		t.Error(err)
	}

	m1 := &Message{}
	s.Deserialize(b, m1)

	if !reflect.DeepEqual(m, m1) {
		t.Fail()
	}
}
