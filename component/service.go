package component

import (
	"errors"
	"reflect"
	"sync"
)

type HandlerMethod struct {
	sync.Mutex // protects counters
	Method     reflect.Method
	Type       reflect.Type
	Raw        bool //Whether the data need to serialize
	numCalls   uint
}

type RemoteMethod struct {
	sync.Mutex // protects counters
	Method     reflect.Method
	Type       reflect.Type
	numCalls   uint
}

type Service struct {
	Name          string                    // name of service
	Rcvr          reflect.Value             // receiver of methods for the service
	Type          reflect.Type              // type of the receiver
	HandlerMethod map[string]*HandlerMethod // registered methods
	RemoteMethod  map[string]*RemoteMethod  // registered methods
}

func (s *Service) ScanHandler() error {
	if s.Name == "" {
		return errors.New("handler.Register: no service name for type " + s.Type.String())
	}
	if !isExported(s.Name) {
		return errors.New("handler.Register: type " + s.Name + " is not exported")
	}

	// Install the methods
	s.HandlerMethod = suitableHandlerMethods(s.Type, true)

	if len(s.HandlerMethod) == 0 {
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

func (s *Service) ScanRemote() error {
	if s.Name == "" {
		return errors.New("handler.Register: no service name for type " + s.Type.String())
	}
	if !isExported(s.Name) {
		return errors.New("handler.Register: type " + s.Name + " is not exported")
	}

	// Install the remote methods
	s.RemoteMethod = suitableRemoteMethods(s.Type, true)
	if len(s.HandlerMethod) == 0 {
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
