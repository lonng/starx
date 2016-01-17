package starx

type Channel struct {
	name           string
	uids           []int
	count          int
	channelServive *ChannelServive // channel service which contain current channel
}

func NewChannel(n string, cs *ChannelServive) *Channel {
	return &Channel{
		name:           n,
		channelServive: cs}
}

func (c *Channel) GetMembers() []int {
	return c.uids
}

func (c *Channel) PushMessageByUids(uids []int, route string, data []byte) {
	netService.Multcast(uids, route, data)
}

func (c *Channel) Broadcast(route string, data []byte) {
	netService.Multcast(c.uids, route, data)
}

func (c *Channel) IsContain(uid int) bool {
	for _, u := range c.uids {
		if u == uid {
			return true
		}
	}

	return false
}

func (c *Channel) Add(uid int) {
	c.uids = append(c.uids, uid)
	c.count++
}

func (c *Channel) Leave(uid int) {
	var temp []int
	for i, u := range c.uids {
		if u == uid {
			temp = append(temp, c.uids[:i]...)
			c.uids = append(temp, c.uids[(i+1):]...)
			break
		}
	}
}

func (c *Channel) LeaveAll() {
	c.uids = make([]int, 0)
}

func (c *Channel) GetCount() int {
	return c.count
}

func (c *Channel) Destroy() {
	delete(c.channelServive.channels, c.name)
}
