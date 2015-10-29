package starx

import (
	"fmt"
	"net"
	"sync"
)

type Session struct {
	Id      int
	Uid     int
	RawConn net.Conn
}

func NewSession(rawConn net.Conn) *Session {
	return &Session{
		Id:      connectionService.getNewSessionUUID(),
		RawConn: rawConn}
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

func (s *SessionService) RegisterSession(session *Session) {
	if _, exists := s.getSessionByAddr(session.RawConn.RemoteAddr().String()); exists {
		Info("session has exists already")
		return
	}
	s.sam.Lock()
	s.sessionAddrMaps[session.RawConn.RemoteAddr().String()] = session
	s.sam.Unlock()
	connectionService.incrementConnCount()
}

func (s *SessionService) RemoveSession(addr string) {
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
