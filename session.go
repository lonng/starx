package starx

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type SessionStatus byte

const (
	_ SessionStatus = iota
	SS_START
	SS_HANDSHAKING
	SS_WORKING
	SS_CLOSED
)

type Session struct {
	Id       int
	Uid      int
	reqId    uint // last requst id
	RawConn  net.Conn
	status   SessionStatus
	lastTime int64
}

func NewSession(rawConn net.Conn) *Session {
	return &Session{
		Id:       connectionService.getNewSessionUUID(),
		RawConn:  rawConn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

func (session *Session) Bind(uid int) {
	if session.Uid > 0 {
		session.Uid = uid
		sessionService.SessionUidMaps[uid] = session
	} else {
		Error("uid invalid")
	}
}

func (session *Session) String() string {
	return fmt.Sprintf("Id: %d, Uid: %d, RemoteAddr: %s",
		session.Id,
		session.Uid,
		session.RawConn.RemoteAddr().String())
}

func (session *Session) heartbeat() {
	Info("session: heartbeat")
	session.lastTime = time.Now().Unix()
}

type SessionService struct {
	sum             sync.RWMutex     // protect SessionUidMaps
	SessionUidMaps  map[int]*Session // uid map sesseion
	sam             sync.RWMutex     // protect sessionAddrMaps
	sessionAddrMaps map[string]*Session
}

func NewSesseionService() *SessionService {
	return &SessionService{
		SessionUidMaps:  make(map[int]*Session),
		sessionAddrMaps: make(map[string]*Session)}
}

func (s *SessionService) RegisterSession(conn net.Conn) *Session {
	if session, exists := s.getSessionByAddr(conn.RemoteAddr().String()); exists {
		Info("session has exists already")
		return session
	}
	session := NewSession(conn)
	s.sam.Lock()
	s.sessionAddrMaps[session.RawConn.RemoteAddr().String()] = session
	s.sam.Unlock()
	connectionService.incrementConnCount()
	return session
}

func (s *SessionService) RemoveSession(session *Session) {
	addr := session.RawConn.RemoteAddr().String()
	if session, exists := s.getSessionByAddr(addr); exists {
		s.sum.Lock()
		s.sam.Lock()
		if session.Uid > 0 {
			delete(s.SessionUidMaps, session.Uid)
		}
		delete(s.sessionAddrMaps, addr)
		s.sam.Unlock()
		s.sum.Unlock()

	} else {
		Info("session has not exists")
	}
	connectionService.decrementConnCount()
}

func (s *SessionService) getSessionByAddr(addr string) (*Session, bool) {
	s.sam.RLock()
	session, exists := s.sessionAddrMaps[addr]
	s.sam.RUnlock()
	return session, exists
}

// Decide whether a remote address session exists
func (s *SessionService) isSessionExists(addr string) bool {
	s.sam.RLock()
	_, exists := s.sessionAddrMaps[addr]
	s.sam.RUnlock()
	return exists
}

func (s *SessionService) GetSessionByUid(uid int) (*Session, bool) {
	s.sum.RLock()
	session, exists := s.SessionUidMaps[uid]
	s.sum.RUnlock()
	return session, exists
}

func (s *SessionService) dumpSessions() {
	s.sam.RLock()
	Info(fmt.Sprintf("current session count: %d", len(s.sessionAddrMaps)))
	for _, ses := range s.sessionAddrMaps {
		Info("session: " + ses.String())
	}
	s.sam.RUnlock()
}
