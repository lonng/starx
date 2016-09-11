package starx

import (
	"github.com/chrislonng/starx/serialize"
	"github.com/chrislonng/starx/serialize/protobuf"
)

// Default serializer
var serializer serialize.Serializer = protobuf.NewProtobufSerializer()

// Customize serializer
func SetSerializer(seri serialize.Serializer) {
	serializer = seri
}
