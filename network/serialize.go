package network

import (
	"github.com/chrislonng/starx/serialize"
	"github.com/chrislonng/starx/serialize/protobuf"
)

// Default serializer
var serializer serialize.Serializer = protobuf.NewProtobufSerializer()

// Customize serializer
func Serializer(seri serialize.Serializer) {
	serializer = seri
}
