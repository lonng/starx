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

func (net *NetService) Broadcast(uids []int, data []byte) {
	for _, uid := range uids {
		net.Send(uid, data)
	}
}
