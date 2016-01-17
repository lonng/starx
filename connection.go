package starx

import (
	"sync"
)

type connectionService struct {
	connCountLock sync.RWMutex // protect connCount
	connCount     int
	sessUIDLock   sync.RWMutex // protect sessUID
	sessUID       uint64
}

func newConnectionService() *connectionService {
	return &connectionService{
		sessUID: 0}
}

func (c *connectionService) incrementConnCount() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount++
}

func (c *connectionService) decrementConnCount() {
	c.connCountLock.Lock()
	defer c.connCountLock.Unlock()
	c.connCount--
}

func (c *connectionService) getCurrentConnCount() int {
	c.connCountLock.RLock()
	defer c.connCountLock.RUnlock()
	return c.connCount
}

func (c *connectionService) getNewSessionUUID() uint64 {
	c.sessUIDLock.Lock()
	defer c.sessUIDLock.Unlock()
	c.sessUID++
	return c.sessUID
}

func (c *connectionService) getCurrentSessionUUID() uint64 {
	c.sessUIDLock.RLock()
	defer c.sessUIDLock.RUnlock()
	return c.sessUID
}
