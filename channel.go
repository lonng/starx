package starx

type Channel struct {
	name           string              // channel name
	uidMap         map[uint64]*Session // uid map to session pointer
	uids           []uint64            // all user ids
	count          int                 // current channel contain user count
	channelServive *channelServive     // channel service which contain current channel
}

func newChannel(n string, cs *channelServive) *Channel {
	return &Channel{
		name:           n,
		channelServive: cs,
		uidMap:         make(map[uint64]*Session)}
}

func (c *Channel) GetMembers() []uint64 {
	return c.uids
}

func (c *Channel) PushMessageByUids(uids []uint64, route string, data []byte) {
	for _, uid := range uids {
		if session, ok := c.uidMap[uid]; ok && session != nil {
			defaultNetService.Push(session, route, data)
		}
	}
}

func (c *Channel) Broadcast(route string, data []byte) {
	for _, session := range c.uidMap {
		defaultNetService.Push(session, route, data)
	}
}

func (c *Channel) IsContain(uid uint64) bool {
	for _, u := range c.uids {
		if u == uid {
			return true
		}
	}
	return false
}

func (c *Channel) Add(session *Session) {
	c.uidMap[session.Uid] = session
	c.uids = append(c.uids, session.Uid)
	c.count++
}

func (c *Channel) Leave(uid uint64) {
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
	c.uids = make([]uint64, 0)
	c.count = 0
}

func (c *Channel) GetCount() int {
	return c.count
}

func (c *Channel) Destroy() {
	delete(c.channelServive.channels, c.name)
}
