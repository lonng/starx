package starx

import (
	"github.com/chrislonng/starx/packet"
	"reflect"
	"testing"
)

func Test1(t *testing.T) {
	if !reflect.DeepEqual(heartbeatPacket, append([]byte{packet.Heartbeat, 0x00, 0x00, 0x00})) {
		t.Error("wrong heartbeat packet")
	}
}
