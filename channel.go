package starx

import (
	"github.com/chrislonng/starx/log"
	"github.com/chrislonng/starx/session"
	"sync"
)

type Channel struct {
	sync.RWMutex
	name           string                     // channel name
	uidMap         map[int64]*session.Session // uid map to session pointer
	uids           []int64                    // all user ids
	count          int                        // current channel contain user count
	channelServive *channelService            // channel service which contain current channel
}

func newChannel(n string, cs *channelService) *Channel {
	return &Channel{
		name:           n,
		channelServive: cs,
		uidMap:         make(map[int64]*session.Session)}
}

func (c *Channel) Members() []int64 {
	c.RLock()
	defer c.RUnlock()

	return c.uids
}

func (c *Channel) PushMessageByUids(uids []int64, route string, data []byte) {
	c.RLock()
	defer c.RUnlock()

	for _, uid := range uids {
		if s, ok := c.uidMap[uid]; ok && s != nil {
			defaultNetService.Push(s, route, data)
		}
	}
}

func (c *Channel) Broadcast(route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	c.RLock()
	defer c.RUnlock()

	for _, s := range c.uidMap {
		err = defaultNetService.Push(s, route, data)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return err
}

func (c *Channel) IsContain(uid int64) bool {
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

func (c *Channel) Leave(uid int64) {
	c.Lock()
	defer c.Unlock()

	var temp []int64
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

	c.uids = make([]int64, 0)
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
