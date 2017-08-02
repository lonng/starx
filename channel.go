package starx

import (
	"sync"

	"github.com/lonnng/starx/log"
	"github.com/lonnng/starx/session"
)

type Channel struct {
	sync.RWMutex
	name    string                     // channel name
	uidMap  map[int64]*session.Session // uid map to session pointer
	members []int64                    // all user ids
}

func newChannel(n string) *Channel {
	return &Channel{
		name:   n,
		uidMap: make(map[int64]*session.Session)}
}

func (c *Channel) Member(uid int64) *session.Session {
	c.RLock()
	defer c.RUnlock()

	return c.uidMap[uid]
}

func (c *Channel) Members() []int64 {
	c.RLock()
	defer c.RUnlock()

	return c.members
}

// Push message to partial client, which filter return true
func (c *Channel) Multicast(route string, v interface{}, filter SessionFilter) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Multicast Route=%s, Data=%+v", route, v)

	c.RLock()
	defer c.RUnlock()

	for _, s := range c.uidMap {
		if !filter(s) {
			continue
		}
		err = transporter.push(s, route, data)
		if err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

// Push message to all client
func (c *Channel) Broadcast(route string, v interface{}) error {
	data, err := serializeOrRaw(v)
	if err != nil {
		return err
	}

	log.Debugf("Type=Broadcast Route=%s, Data=%+v", route, v)

	c.RLock()
	defer c.RUnlock()

	for _, s := range c.uidMap {
		err = transporter.push(s, route, data)
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
	c.LeaveAll()
}
