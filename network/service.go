package network

import (
	"reflect"
	"sync"
)

type handlerMethod struct {
	sync.Mutex // protects counters
	method     reflect.Method
	dataType   reflect.Type
	raw        bool //Whether the data need to serialize
	numCalls   uint
}

type remoteMethod struct {
	sync.Mutex // protects counters
	method     reflect.Method
	dataType   reflect.Type
	numCalls   uint
}

type service struct {
	name          string                    // name of service
	rcvr          reflect.Value             // receiver of methods for the service
	typ           reflect.Type              // type of the receiver
	handlerMethod map[string]*handlerMethod // registered methods
	remoteMethod  map[string]*remoteMethod  // registered methods
}

func (m *handlerMethod) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (m *remoteMethod) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}
