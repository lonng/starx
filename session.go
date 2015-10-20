package mello

import (
	"net"
)

type Session struct {
	Id      int64
	Uid     int64
	RawConn net.Conn
}

func (session *Session) Bind(uid int64) {
	if session.Uid > 0 {

	}
	session.Uid = uid
}

type MelloSessionService struct {
	Sessions       []*Session         // all sessions
	SessionUidMaps map[int64]*Session // uid map sesseion
}

func NewSesseionService() *MelloSessionService {
	return &MelloSessionService{}
}
