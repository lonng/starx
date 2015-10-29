package starx

import (
	"net"
)

type Session struct {
	Id      int64
	Uid     int64
	RawConn net.Conn
}

func NewSession(rawConn net.Conn) *Session {
	return &Session{RawConn: rawConn}
}

func (session *Session) Bind(uid int64) {
	if session.Uid > 0 {

	}
	session.Uid = uid
}

type SessionService struct {
	Sessions       []*Session         // all sessions
	SessionUidMaps map[int64]*Session // uid map sesseion
}

func NewSesseionService() *SessionService {
	return &SessionService{}
}
