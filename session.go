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

// This session type as argument pass to Handler method, is a proxy session
// for frontend session in frontend server or backend session in backend
// server, correspond frontend session or backend session id as a field
// will be store in type instance
type Session struct {
	Id                int
	Uid               int
	reqId             uint // last requst id
	status            SessionStatus
	lastTime          int64
	frontendSessionId uint64 // current session correspond frontend session, only in frontend server
	backendSessionId  uint64 // current session correspond backend session, only in backend server
}

// Session for frontend server, used for store raw socket information
type frontendSession struct {
	id       uint64
	socket   net.Conn
	status   SessionStatus
	lastTime int64 // last heartbeat unix time stamp
}

// Session for backend server, used for store raw socket information
type backendSession struct {
	id         uint64
	socket     net.Conn
	status     SessionStatus
	sessionMap map[uint64]*Session
	lastTime   int64 // last heartbeat unix time stamp
}

// Create new session instance
func newSession(rawConn net.Conn) *Session {
	return &Session{
		Id:       connectionService.getNewSessionUUID(),
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

// Create new frontend session instance
func newFrontendSession(id uint64, conn net.Conn) *frontendSession {
	return &frontendSession{
		id:       id,
		socket:   conn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

// Create new backend session instance
func newBackendSession(id uint64, conn net.Conn) *backendSession {
	return &backendSession{
		id:       id,
		socket:   conn,
		status:   SS_START,
		lastTime: time.Now().Unix()}
}

// Session send packet data
func (session *Session) Send(data []byte) {
	Net.send(session, data)
}

// Push message to session
func (session *Session) Push(route string, data[]byte) {
	Net.Push(session, route, data)
}

// Response message to session
func (session *Session) Response(data []byte){
	Net.Response(session, data)
}

// Implement Stringer interface
func (fs *frontendSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		fs.id,
		fs.socket.RemoteAddr().String(),
		fs.lastTime)
}

// Implement Stringer interface
func (bs *backendSession) String() string {
	return fmt.Sprintf("id: %d, remote address: %s, last time: %d",
		bs.id,
		bs.socket.RemoteAddr().String(),
		bs.lastTime)
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
	session := newSession(conn)
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
