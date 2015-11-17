/*
 消息发送
*/
package starx

import (
	"fmt"
)

type NetService struct {
}

func NewNetService() *NetService {
	return &NetService{}
}

func (net *NetService) Send(uid int, data []byte) {
	if session, ok := sessionService.GetSessionByUid(uid); ok {
		_, err := session.RawConn.Write(data)
		if err != nil {
			Info(err.Error())
		}
	} else {
		Info(fmt.Sprintf("uid: %d not found", uid))
	}
}

func (net *NetService) SendToSession(session *Session, data []byte) {
	session.RawConn.Write(pack(PACKET_DATA, data))
}

func (net *NetService) PushToSession(session *Session, route string, data []byte) {
	m := encodeMessage(&Message{Type: MessageType(MT_PUSH), Route: route, Body: data})
	go session.RawConn.Write(pack(PacketType(PACKET_DATA), m))
}

func (net *NetService) ResponseToSession(session *Session, data []byte) {
	if session.reqId > 0 {
		m := encodeMessage(&Message{Type: MessageType(MT_RESPONSE), ID: session.reqId, Body: data})
		go session.RawConn.Write(pack(PacketType(PACKET_DATA), m))
	}
}

func (net *NetService) Multicast(uids []int, data []byte) {
	for _, uid := range uids {
		net.Send(uid, data)
	}
}

func (net *NetService) Broadcast(data []byte) {
	for _, session := range sessionService.sessionAddrMaps {
		_, err := session.RawConn.Write(data)
		if err != nil {
			Info(err.Error())
		}
	}
}
