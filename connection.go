package starx

import (
	"sync"
)

type ConnectionService struct {
	countLock sync.RWMutex // protect connCount
	connCount int
	uidLock   sync.RWMutex // protect sessionID
	sessionID uint64
}

func newConnectionService() *ConnectionService {
	return &ConnectionService{
		sessionID: 0}
}

func (c *ConnectionService) Increment() {
	c.countLock.Lock()
	defer c.countLock.Unlock()
	c.connCount++
}

func (c *ConnectionService) Decrement() {
	c.countLock.Lock()
	defer c.countLock.Unlock()
	c.connCount--
}

func (c *ConnectionService) Count() int {
	c.countLock.RLock()
	defer c.countLock.RUnlock()
	return c.connCount
}

func (c *ConnectionService) NewSessionUUID() uint64 {
	c.uidLock.Lock()
	defer c.uidLock.Unlock()
	c.sessionID++
	return c.sessionID
}

func (c *ConnectionService) SessionUUID() uint64 {
	c.uidLock.RLock()
	defer c.uidLock.RUnlock()
	return c.sessionID
}
