package session

import (
	"errors"
	"fmt"
	"time"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/service"
)

type Entity interface {
	ID() uint64
	Send([]byte) error
	Push(session *Session, route string, v interface{}) error
	Response(session *Session, v interface{}) error
	AsyncCall(session *Session, route string, args ...interface{}) error
	Call(session *Session, route string, args ...interface{}) ([]byte, error)
}

var (
	ErrIllegalUID = errors.New("illegal uid")
)

// This session type as argument pass to Handler method, is a proxy session
// for frontend session in frontend server or backend session in backend
// server, correspond frontend session or backend session id as a field
// will be store in type instance
//
// This is user sessions, not contain raw sockets information
type Session struct {
	Id       uint64                 // session global uniqe id
	Uid      uint64                 // binding user id
	Entity   Entity                 // raw session id, frontendSession in frontend server, or backendSession in backend server
	LastID   uint                   // last request id
	data     map[string]interface{} // session data store
	lastTime int64                  // last heartbeat time
}

// Create new session instance
func NewSession(entity Entity) *Session {
	return &Session{
		Id:       service.Connections.NewSessionUUID(),
		Entity:   entity,
		data:     make(map[string]interface{}),
		lastTime: time.Now().Unix(),
	}
}

// Session send packet data
func (s *Session) Send(data []byte) error {
	return s.Entity.Send(data)
}

// Push message to session
func (s *Session) Push(route string, v interface{}) error {
	return s.Entity.Push(s, route, v)
}

// Response message to session
func (s *Session) Response(v interface{}) error {
	return s.Entity.Response(s, v)
}

func (s *Session) Bind(uid uint64) error {
	if uid < 1 {
		log.Error("uid invalid: %d", uid)
		return ErrIllegalUID
	}
	s.Uid = uid
	return nil
}

func (s *Session) String() string {
	return fmt.Sprintf("Id: %d, Uid: %d", s.Id, s.Uid)
}

func (s *Session) AsyncCall(route string, args ...interface{}) error {
	return s.Entity.AsyncCall(s, route, args...)
}

func (s *Session) Call(route string, args ...interface{}) ([]byte, error) {
	return s.Entity.Call(s, route, args...)
}

// Sync session setting to frontend server
func (s *Session) Sync(string) {
	//TODO
	//synchronize session setting field to frontend server
}

// Sync all settings to frontend server
func (s *Session) SyncAll() {
}

// TODO: session data setting interface
// ????
