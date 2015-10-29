package starx

import (
	"sync"
)

type ConnectionService struct {
	connCountLock sync.RWMutex // protect connCount
	connCount     int
	sessUUIDLock  sync.RWMutex // protect sessUUID
	sessUUID      int
}

func NewConnectionService() *ConnectionService {
	return &ConnectionService{
		sessUUID: 0}
}

func (c *ConnectionService) incrementConnCount() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount++
}

func (c *ConnectionService) decrementConnCount() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount--
}

func (c *ConnectionService) getCurrentConnCount() int {
	c.connCountLock.RLock()
	defer c.connCountLock.RUnlock()
	return c.connCount
}

func (c *ConnectionService) getNewSessionUUID() int {
	c.sessUUIDLock.Lock()
	defer c.sessUUIDLock.Unlock()
	c.sessUUID++
	return c.sessUUID
}

func (c *ConnectionService) getCurrentSessionUUID() int {
	c.sessUUIDLock.RLock()
	defer c.sessUUIDLock.RUnlock()
	return c.sessUUID
}
