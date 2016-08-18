package network

import (
	"github.com/chrislonng/starx/session"
	"sync"
)

type Channel struct {
	sync.RWMutex
	name           string                      // channel name
	uidMap         map[uint64]*session.Session // uid map to session pointer
	uids           []uint64                    // all user ids
	count          int                         // current channel contain user count
	channelServive *channelService             // channel service which contain current channel
}

func newChannel(n string, cs *channelService) *Channel {
	return &Channel{
		name:           n,
		channelServive: cs,
		uidMap:         make(map[uint64]*session.Session)}
}

func (c *Channel) Members() []uint64 {
	c.RLock()
	defer c.RUnlock()

	return c.uids
}

func (c *Channel) PushMessageByUids(uids []uint64, route string, data []byte) {
	c.RLock()
	defer c.RUnlock()

	for _, uid := range uids {
		if s, ok := c.uidMap[uid]; ok && s != nil {
			defaultNetService.Push(s, route, data)
		}
	}
}

func (c *Channel) Broadcast(route string, data []byte) {
	c.RLock()
	defer c.RUnlock()

	for _, s := range c.uidMap {
		defaultNetService.Push(s, route, data)
	}
}

func (c *Channel) IsContain(uid uint64) bool {
	c.RLock()
	defer c.RUnlock()

	for _, u := range c.uids {
		if u == uid {
			return true
		}
	}
	return false
}

func (c *Channel) Add(session *session.Session) {
	c.Lock()
	defer c.Unlock()

	c.uidMap[session.Uid] = session
	c.uids = append(c.uids, session.Uid)
	c.count++
}

func (c *Channel) Leave(uid uint64) {
	c.Lock()
	defer c.Unlock()

	var temp []uint64
	for i, u := range c.uids {
		if u == uid {
			temp = append(temp, c.uids[:i]...)
			c.uids = append(temp, c.uids[(i+1):]...)
			c.count--
			break
		}
	}
}

func (c *Channel) LeaveAll() {
	c.Lock()
	defer c.Unlock()

	c.uids = make([]uint64, 0)
	c.count = 0
}

func (c *Channel) Count() int {
	c.RLock()
	defer c.RUnlock()

	return c.count
}

func (c *Channel) Destroy() {
	c.channelServive.Lock()
	defer c.channelServive.Unlock()

	delete(c.channelServive.channels, c.name)
}
