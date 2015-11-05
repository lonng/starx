package starx

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type HandlerService struct{}

func NewHandler() *HandlerService {
	return &HandlerService{}
}

func (handler *HandlerService) Handle(conn net.Conn) {
	defer conn.Close()
	session := sessionService.RegisterSession(conn)
	sessionService.dumpSessions()
	tmp := make([]byte, 0) //保存截断数据
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			Info("session closed(" + err.Error() + ")")
			session.status = SS_CLOSED
			sessionService.RemoveSession(session)
			sessionService.dumpSessions()
			break
		}

		// decode packet
		var pkg *Packet
		pkg, tmp = unpack(append(tmp, buf[:n]...))
		if pkg != nil {
			switch pkg.Type {
			case PACKET_HANDSHAKE:
				{
					session.status = SS_HANDSHAKING
					Info(pkg.String())
					data, err := json.Marshal(map[string]interface{}{"code": 200, "sys": map[string]float64{"heartbeat": heartbeatInternal.Seconds()}})
					if err != nil {
						Info(err.Error())
					}
					conn.Write(pack(PACKET_HANDSHAKE, data))
				}
			case PACKET_HANDSHAKE_ACK:
				{
					session.status = SS_WORKING
				}
			case PACKET_HEARTBEAT:
				{
					session.heartbeat()
				}
			case PACKET_DATA:
				{
					session.heartbeat()
					msg := decodeMessage(pkg.Body)
					if msg != nil {
						App.MessageChan <- msg
						if msg.Type == MT_REQUEST {
							time.AfterFunc(time.Second*5, func() {
								Net.SendToSession(session, encodeMessage(&Message{Type: MT_RESPONSE, ID: msg.ID, Body: []byte(`{"code": "test message"}`)}))
							})
						} else {
							time.AfterFunc(time.Second*5, func() {
								Net.SendToSession(session, encodeMessage(&Message{Type: MT_PUSH, Route: "mailSystem.New", Body: []byte(`{"code": "test message"}`)}))
							})
						}
					}
				}
			}
		}
	}
}

func encodeMessage(m *Message) []byte {
	temp := make([]byte, 0)
	flag := byte(m.Type) << 1
	if m.isCompress {
		flag |= 0x01
	}
	temp = append(temp, flag)
	// response message
	if m.Type == MT_RESPONSE {
		if m.ID > 0 {
			n := m.ID
			for {
				b := byte(n % 128)
				n >>= 7
				if n != 0 {
					temp = append(temp, b+128)
				} else {
					temp = append(temp, b)
					break
				}
			}
			fmt.Println("%+v", temp)
		}
	} else if m.Type == MT_PUSH {
		if m.isCompress {
			temp = append(temp, byte((m.RouteCode>>8)&0xFF))
			temp = append(temp, byte(m.RouteCode&0xFF))
		} else {
			temp = append(temp, byte(len(m.Route)))
			temp = append(temp, []byte(m.Route)...)
		}
	} else {
		Error("wrong message type")
	}
	temp = append(temp, m.Body...)
	return temp
}

func decodeMessage(data []byte) *Message {
	// filter invalid message
	if len(data) <= 3 {
		Info("invalid message")
		return nil
	}
	msg := NewMessage()
	flag := data[0]
	// set offset to 1, because 1st byte will always be flag
	offset := 1
	msg.Type = MessageType((flag >> 1) & MSG_TYPE_MASK)
	if msg.Type == MT_REQUEST {
		id := 0
		// little end byte order
		// WARNING: must can be stored in 64 bits integer
		for i := offset; i < len(data); i++ {
			b := data[i]
			id += (int(b&0x7F) << uint(7*(i-offset)))
			if b < 128 {
				offset = i + 1
				break
			}
		}
		msg.ID = id
	}
	if flag&MSG_ROUTE_COMPRESS_MASK == 1 {
		msg.isCompress = true
		msg.RouteCode = bytesToInt(data[offset:(offset + 2)])
		offset += 2
	} else {
		msg.isCompress = false
		rl := data[offset]
		offset += 1
		msg.Route = string(data[offset:(offset + int(rl))])
		offset += int(rl)
	}
	msg.Body = data[offset:]
	return msg
}

func (handler *HandlerService) Register(rcvr HandlerComponent) {
	Info(fmt.Sprintf("Register Handler: %s", rcvr))
	rcvr.Setup()
}
