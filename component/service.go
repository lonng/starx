package component

import (
	"errors"
	"reflect"
	"sync"
)

type HandlerMethod struct {
	sync.Mutex
	Method   reflect.Method
	Type     reflect.Type
	Raw      bool //Whether the data need to serialize
	numCalls uint
}

type RemoteMethod struct {
	sync.Mutex
	Method   reflect.Method
	Type     reflect.Type
	numCalls uint
}

type Service struct {
	Name           string                    // name of service
	Rcvr           reflect.Value             // receiver of methods for the service
	Type           reflect.Type              // type of the receiver
	HandlerMethods map[string]*HandlerMethod // registered methods
	RemoteMethods  map[string]*RemoteMethod  // registered methods
}

// Register publishes in the service the set of methods of the
// receiver value that satisfy the following conditions:
// - exported method of exported type
// - two arguments, both of exported type
// - the first argument is *session.Session
// - the second argument is []byte or a pointer
func (s *Service) ScanHandler() error {
	if s.Name == "" {
		return errors.New("handler.Register: no service name for type " + s.Type.String())
	}
	if !isExported(s.Name) {
		return errors.New("handler.Register: type " + s.Name + " is not exported")
	}

	// Install the methods
	s.HandlerMethods = suitableHandlerMethods(s.Type, true)

	if len(s.HandlerMethods) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableHandlerMethods(reflect.PtrTo(s.Type), false)
		if len(method) != 0 {
			str = "handler.Register: type " + s.Name + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "handler.Register: type " + s.Name + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	return nil
}

// Register publishes in the service the set of methods of the
// receiver value that satisfy the following conditions:
// - exported method of exported type
// - two return value, the last one must be error
func (s *Service) ScanRemote() error {
	if s.Name == "" {
		return errors.New("handler.Register: no service name for type " + s.Type.String())
	}
	if !isExported(s.Name) {
		return errors.New("handler.Register: type " + s.Name + " is not exported")
	}

	// Install the remote methods
	s.RemoteMethods = suitableRemoteMethods(s.Type, true)
	if len(s.HandlerMethods) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableRemoteMethods(reflect.PtrTo(s.Type), false)
		if len(method) != 0 {
			str = "remote.Register: type " + s.Name + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "remote.Register: type " + s.Name + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	return nil
}

func (m *HandlerMethod) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (m *RemoteMethod) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}
