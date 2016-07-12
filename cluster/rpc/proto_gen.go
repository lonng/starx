package rpc

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *Request) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zxvk uint32
	zxvk, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zxvk > 0 {
		zxvk--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ServiceMethod":
			z.ServiceMethod, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Seq":
			z.Seq, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Sid":
			z.Sid, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Args":
			z.Data, err = dc.ReadBytes(z.Data)
			if err != nil {
				return
			}
		case "Kind":
			{
				var zbzg byte
				zbzg, err = dc.ReadByte()
				z.Kind = RpcKind(zbzg)
			}
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Request) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "ServiceMethod"
	err = en.Append(0x85, 0xad, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ServiceMethod)
	if err != nil {
		return
	}
	// write "Seq"
	err = en.Append(0xa3, 0x53, 0x65, 0x71)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Seq)
	if err != nil {
		return
	}
	// write "Sid"
	err = en.Append(0xa3, 0x53, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Sid)
	if err != nil {
		return
	}
	// write "Args"
	err = en.Append(0xa4, 0x41, 0x72, 0x67, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.Data)
	if err != nil {
		return
	}
	// write "Kind"
	err = en.Append(0xa4, 0x4b, 0x69, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteByte(byte(z.Kind))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Request) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "ServiceMethod"
	o = append(o, 0x85, 0xad, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	o = msgp.AppendString(o, z.ServiceMethod)
	// string "Seq"
	o = append(o, 0xa3, 0x53, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Seq)
	// string "Sid"
	o = append(o, 0xa3, 0x53, 0x69, 0x64)
	o = msgp.AppendUint64(o, z.Sid)
	// string "Args"
	o = append(o, 0xa4, 0x41, 0x72, 0x67, 0x73)
	o = msgp.AppendBytes(o, z.Data)
	// string "Kind"
	o = append(o, 0xa4, 0x4b, 0x69, 0x6e, 0x64)
	o = msgp.AppendByte(o, byte(z.Kind))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Request) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zbai uint32
	zbai, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zbai > 0 {
		zbai--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "ServiceMethod":
			z.ServiceMethod, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Seq":
			z.Seq, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Sid":
			z.Sid, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Args":
			z.Data, bts, err = msgp.ReadBytesBytes(bts, z.Data)
			if err != nil {
				return
			}
		case "Kind":
			{
				var zcmr byte
				zcmr, bts, err = msgp.ReadByteBytes(bts)
				z.Kind = RpcKind(zcmr)
			}
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Request) Msgsize() (s int) {
	s = 1 + 14 + msgp.StringPrefixSize + len(z.ServiceMethod) + 4 + msgp.Uint64Size + 4 + msgp.Uint64Size + 5 + msgp.BytesPrefixSize + len(z.Data) + 5 + msgp.ByteSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *Response) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zajw uint32
	zajw, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zajw > 0 {
		zajw--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Kind":
			{
				var zwht byte
				zwht, err = dc.ReadByte()
				z.Kind = ResponseKind(zwht)
			}
			if err != nil {
				return
			}
		case "ServiceMethod":
			z.ServiceMethod, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Seq":
			z.Seq, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Sid":
			z.Sid, err = dc.ReadUint64()
			if err != nil {
				return
			}
		case "Data":
			z.Data, err = dc.ReadBytes(z.Data)
			if err != nil {
				return
			}
		case "Error":
			z.Error, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Route":
			z.Route, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *Response) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 7
	// write "Kind"
	err = en.Append(0x87, 0xa4, 0x4b, 0x69, 0x6e, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteByte(byte(z.Kind))
	if err != nil {
		return
	}
	// write "ServiceMethod"
	err = en.Append(0xad, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ServiceMethod)
	if err != nil {
		return
	}
	// write "Seq"
	err = en.Append(0xa3, 0x53, 0x65, 0x71)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Seq)
	if err != nil {
		return
	}
	// write "Sid"
	err = en.Append(0xa3, 0x53, 0x69, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteUint64(z.Sid)
	if err != nil {
		return
	}
	// write "Data"
	err = en.Append(0xa4, 0x44, 0x61, 0x74, 0x61)
	if err != nil {
		return err
	}
	err = en.WriteBytes(z.Data)
	if err != nil {
		return
	}
	// write "Error"
	err = en.Append(0xa5, 0x45, 0x72, 0x72, 0x6f, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Error)
	if err != nil {
		return
	}
	// write "Route"
	err = en.Append(0xa5, 0x52, 0x6f, 0x75, 0x74, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Route)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *Response) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 7
	// string "Kind"
	o = append(o, 0x87, 0xa4, 0x4b, 0x69, 0x6e, 0x64)
	o = msgp.AppendByte(o, byte(z.Kind))
	// string "ServiceMethod"
	o = append(o, 0xad, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	o = msgp.AppendString(o, z.ServiceMethod)
	// string "Seq"
	o = append(o, 0xa3, 0x53, 0x65, 0x71)
	o = msgp.AppendUint64(o, z.Seq)
	// string "Sid"
	o = append(o, 0xa3, 0x53, 0x69, 0x64)
	o = msgp.AppendUint64(o, z.Sid)
	// string "Data"
	o = append(o, 0xa4, 0x44, 0x61, 0x74, 0x61)
	o = msgp.AppendBytes(o, z.Data)
	// string "Error"
	o = append(o, 0xa5, 0x45, 0x72, 0x72, 0x6f, 0x72)
	o = msgp.AppendString(o, z.Error)
	// string "Route"
	o = append(o, 0xa5, 0x52, 0x6f, 0x75, 0x74, 0x65)
	o = msgp.AppendString(o, z.Route)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Response) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zhct uint32
	zhct, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zhct > 0 {
		zhct--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "Kind":
			{
				var zcua byte
				zcua, bts, err = msgp.ReadByteBytes(bts)
				z.Kind = ResponseKind(zcua)
			}
			if err != nil {
				return
			}
		case "ServiceMethod":
			z.ServiceMethod, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Seq":
			z.Seq, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Sid":
			z.Sid, bts, err = msgp.ReadUint64Bytes(bts)
			if err != nil {
				return
			}
		case "Data":
			z.Data, bts, err = msgp.ReadBytesBytes(bts, z.Data)
			if err != nil {
				return
			}
		case "Error":
			z.Error, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Route":
			z.Route, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

func (z *Response) Msgsize() (s int) {
	s = 1 + 5 + msgp.ByteSize + 14 + msgp.StringPrefixSize + len(z.ServiceMethod) + 4 + msgp.Uint64Size + 4 + msgp.Uint64Size + 5 + msgp.BytesPrefixSize + len(z.Data) + 6 + msgp.StringPrefixSize + len(z.Error) + 6 + msgp.StringPrefixSize + len(z.Route)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *ResponseKind) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zxhx byte
		zxhx, err = dc.ReadByte()
		(*z) = ResponseKind(zxhx)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z ResponseKind) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteByte(byte(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z ResponseKind) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendByte(o, byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *ResponseKind) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zlqf byte
		zlqf, bts, err = msgp.ReadByteBytes(bts)
		(*z) = ResponseKind(zlqf)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z ResponseKind) Msgsize() (s int) {
	s = msgp.ByteSize
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RpcKind) DecodeMsg(dc *msgp.Reader) (err error) {
	{
		var zdaf byte
		zdaf, err = dc.ReadByte()
		(*z) = RpcKind(zdaf)
	}
	if err != nil {
		return
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RpcKind) EncodeMsg(en *msgp.Writer) (err error) {
	err = en.WriteByte(byte(z))
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RpcKind) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	o = msgp.AppendByte(o, byte(z))
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RpcKind) UnmarshalMsg(bts []byte) (o []byte, err error) {
	{
		var zpks byte
		zpks, bts, err = msgp.ReadByteBytes(bts)
		(*z) = RpcKind(zpks)
	}
	if err != nil {
		return
	}
	o = bts
	return
}

func (z RpcKind) Msgsize() (s int) {
	s = msgp.ByteSize
	return
}
