package service

import (
	"sync"
)

var Connections = NewConnectionService()

type connectionService struct {
	countLock sync.RWMutex // protect connCount
	connCount int
	uidLock   sync.RWMutex // protect sessionID
	sessionID int64
}

func NewConnectionService() *connectionService {
	return &connectionService{sessionID: 0}
}

func (c *connectionService) Increment() {
	c.countLock.Lock()
	defer c.countLock.Unlock()
	c.connCount++
}

func (c *connectionService) Decrement() {
	c.countLock.Lock()
	defer c.countLock.Unlock()
	c.connCount--
}

func (c *connectionService) Count() int {
	c.countLock.RLock()
	defer c.countLock.RUnlock()
	return c.connCount
}

func (c *connectionService) NewSessionUUID() int64 {
	c.uidLock.Lock()
	defer c.uidLock.Unlock()
	c.sessionID++
	return c.sessionID
}

func (c *connectionService) SessionUUID() int64 {
	c.uidLock.RLock()
	defer c.uidLock.RUnlock()
	return c.sessionID
}
