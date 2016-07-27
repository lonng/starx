package session

import (
	"errors"
	"time"

	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/service"
	"reflect"
)

type NetworkEntity interface {
	ID() uint64
	Send([]byte) error
	Push(session *Session, route string, v interface{}) error
	Response(session *Session, v interface{}) error
	Call(session *Session, route string, reply interface{}, args ...interface{}) error
	Sync(map[string]interface{}) error
}

var (
	ErrIllegalUID       = errors.New("illegal uid")
	ErrKeyNotFound      = errors.New("current session does not contain key")
	ErrWrongValueType   = errors.New("current key has different data type")
	ErrReplyShouldBePtr = errors.New("reply should be a pointer")
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
	Entity   NetworkEntity          // raw session id, agent in frontend server, or acceptor in backend server
	LastID   uint                   // last request id
	data     map[string]interface{} // session data store
	lastTime int64                  // last heartbeat time
}

// Create new session instance
func NewSession(entity NetworkEntity) *Session {
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

func (s *Session) Call(route string, reply interface{}, args ...interface{}) error {
	if reflect.TypeOf(reply).Kind() != reflect.Ptr {
		return ErrReplyShouldBePtr
	}
	return s.Entity.Call(s, route, reply, args...)
}

// Sync session setting to frontend server
func (s *Session) Sync(key string) error {
	v, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}
	return s.Entity.Sync(map[string]interface{}{
		key: v,
	})
}

// Sync all settings to frontend server
func (s *Session) SyncAll() error {
	if len(s.data) < 1 {
		log.Warn("current session did not contain any data")
		return nil
	}
	return s.Entity.Sync(s.data)
}

func (s *Session) Set(key string, value interface{}) {
	s.data[key] = value
}

func (s *Session) Int(key string) (int, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(int)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Int8(key string) (int8, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(int8)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Int16(key string) (int16, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(int16)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Int32(key string) (int32, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(int32)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Int64(key string) (int64, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(int64)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Uint(key string) (uint, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(uint)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Uint8(key string) (uint8, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(uint8)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Uint16(key string) (uint16, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(uint16)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Uint32(key string) (uint32, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(uint32)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Uint64(key string) (uint64, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(uint64)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Float32(key string) (float32, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(float32)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Float64(key string) (float64, error) {
	v, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	value, ok := v.(float64)
	if !ok {
		return 0, ErrWrongValueType
	}
	return value, nil
}

func (s *Session) String(key string) (string, error) {
	v, ok := s.data[key]
	if !ok {
		return "", ErrKeyNotFound
	}

	value, ok := v.(string)
	if !ok {
		return "", ErrWrongValueType
	}
	return value, nil
}

func (s *Session) Value(key string) (interface{}, error) {
	v, ok := s.data[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return v, nil
}
