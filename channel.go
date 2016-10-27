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
	members        []int64                    // all user ids
	channelService *channelService            // channel service which contain current channel
}

func newChannel(n string, cs *channelService) *Channel {
	return &Channel{
		name:           n,
		channelService: cs,
		uidMap:         make(map[int64]*session.Session)}
}

func (c *Channel) Member(uid int64) *session.Session {
	c.RLock()
	defer c.RUnlock()

	return c.members[uid]
}

func (c *Channel) Members() []int64 {
	c.RLock()
	defer c.RUnlock()

	return c.members
}

func (c *Channel) Multicast(uids []int64, route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	c.RLock()
	defer c.RUnlock()

	for _, uid := range uids {
		if s, ok := c.uidMap[uid]; ok && s != nil {
			defaultNetService.push(s, route, data)
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
		err = defaultNetService.push(s, route, data)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return err
}

func (c *Channel) IsContain(uid int64) bool {
	c.RLock()
	defer c.RUnlock()

	if _, ok := c.uidMap[uid]; ok {
		return true
	}

	return false
}

func (c *Channel) Add(session *session.Session) {
	c.Lock()
	defer c.Unlock()

	c.uidMap[session.Uid] = session
	c.members = append(c.members, session.Uid)
}

func (c *Channel) Leave(uid int64) {
	if !c.IsContain(uid) {
		return
	}

	c.Lock()
	defer c.Unlock()

	var temp []int64
	for i, u := range c.members {
		if u == uid {
			temp = append(temp, c.members[:i]...)
			c.members = append(temp, c.members[(i+1):]...)
			break
		}
	}
	delete(c.uidMap, uid)
}

func (c *Channel) LeaveAll() {
	c.Lock()
	defer c.Unlock()

	c.uidMap = make(map[int64]*session.Session)
	c.members = make([]int64, 0)
}

func (c *Channel) Count() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.uidMap)
}

func (c *Channel) Destroy() {
	c.channelService.Lock()
	defer c.channelService.Unlock()

	delete(c.channelService.channels, c.name)
}
