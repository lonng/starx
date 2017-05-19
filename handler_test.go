package starx

import (
	"reflect"
	"testing"

	"github.com/chrislonng/starx/cluster"
	"github.com/chrislonng/starx/component"
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/message"
	"github.com/chrislonng/starx/serialize/json"
	"github.com/chrislonng/starx/serialize/protobuf"
	"github.com/chrislonng/starx/session"
	"github.com/golang/protobuf/proto"
	"golang.org/x/tools/go/gcimporter15/testdata"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.LevelClose)
	App.Master = &cluster.ServerConfig{
		Type:        "test",
		Id:          "test-1",
		Host:        "127.0.0.1",
		Port:        12305,
		IsFrontend:  false,
		IsMaster:    true,
		IsWebsocket: false,
	}

	App.Config = &cluster.ServerConfig{
		Type:        "test",
		Id:          "test-1",
		Host:        "127.0.0.1",
		Port:        12305,
		IsFrontend:  false,
		IsMaster:    false,
		IsWebsocket: false,
	}

	m.Run()
}

func BenchmarkPointerReflectNewValue(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(&T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t.Elem())
	}

	b.ReportAllocs()
}

func BenchmarkPointerReflectNewInterface(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(&T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t.Elem()).Interface()
	}

	b.ReportAllocs()
}
func BenchmarkReflectNewValue(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t)
	}

	b.ReportAllocs()
}

func BenchmarkReflectNewInterface(b *testing.B) {
	type T struct {
		Code    int
		Message string
		Payload string
	}

	t := reflect.TypeOf(T{})

	for i := 0; i < b.N; i++ {
		reflect.New(t).Interface()
	}

	b.ReportAllocs()
}

// Test types
type (
	TestComp struct {
		component.Base
	}

	JsonMessage struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}

	ProtoMessage struct {
		Data *string `protobuf:"bytes,1,name=data"`
	}
)

func (m *ProtoMessage) Reset()         { *m = ProtoMessage{} }
func (m *ProtoMessage) String() string { return proto.CompactTextString(m) }
func (*ProtoMessage) ProtoMessage()    {}

func (t *TestComp) HandleJson(s *session.Session, m *JsonMessage) error {
	return nil
}

func (t *TestComp) HandleProto(s *session.Session, m *ProtoMessage) error {
	return nil
}

func TestHandlerCallJSON(t *testing.T) {
	SetSerializer(json.NewSerializer())
	handler.register(&TestComp{})

	m := JsonMessage{Code: 1, Data: "hello world"}
	data, err := serializeOrRaw(m)
	if err != nil {
		t.Fail()
	}

	msg := message.New()
	msg.Route = "TestComp.HandleJson"
	msg.Type = message.Request
	msg.Data = data

	s := session.NewSession(nil)

	handler.processMessage(s, msg)
}

func TestHandlerCallProtobuf(t *testing.T) {
	SetSerializer(protobuf.NewSerializer())
	handler.register(&TestComp{})

	m := &ProtoMessage{Data: proto.String("hello world")}
	data, err := serializeOrRaw(m)
	if err != nil {
		t.Error(err)
	}

	msg := message.New()
	msg.Route = "TestComp.HandleProto"
	msg.Type = message.Request
	msg.Data = data

	s := session.NewSession(nil)

	handler.processMessage(s, msg)
}

func BenchmarkHandlerCallJSON(b *testing.B) {
	SetSerializer(json.NewSerializer())
	handler.register(&TestComp{})

	m := JsonMessage{Code: 1, Data: "hello world"}
	data, err := serializeOrRaw(m)
	if err != nil {
		b.Fail()
	}

	msg := message.New()
	msg.Route = "TestComp.HandleJson"
	msg.Type = message.Request
	msg.Data = data

	s := session.NewSession(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.processMessage(s, msg)
	}

	b.ReportAllocs()
}

func BenchmarkHandlerCallProtobuf(b *testing.B) {
	SetSerializer(protobuf.NewSerializer())
	handler.register(&TestComp{})

	m := &ProtoMessage{Data: proto.String("hello world")}
	data, err := serializeOrRaw(m)
	if err != nil {
		b.Fail()
	}

	msg := message.New()
	msg.Route = "TestComp.HandleProto"
	msg.Type = message.Request
	msg.Data = data

	s := session.NewSession(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.processMessage(s, msg)
	}
	b.ReportAllocs()
}
