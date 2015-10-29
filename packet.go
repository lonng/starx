package starx

func Packet(t ProtocolType, data []byte) []byte {
	var buf []byte
	return append(append(append(buf, byte(t)), IntToBytes(len(data))...), data...)
}

func IntToBytes(n int) []byte {
	var buf []byte
	return append(append(buf, byte((n>>8)&0xFF)), byte(n&0xFF))
}
