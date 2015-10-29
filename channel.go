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

func (c *Channel) PushMessageByUids(uids []int, data []byte) {
	for _, uid := range uids {
		if c.IsContain(uid) {
			Net.Send(uid, data)
		}
	}
}

func (c *Channel) Broadcast(data []byte) {
	for _, uid := range c.uids {
		Net.Send(uid, data)
	}
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

type ChannelServive struct {
	channels map[string]*Channel // all server channels
}

func NewChannelServive() *ChannelServive {
	return &ChannelServive{}
}

func (c *ChannelServive) NewChannel(name string) *Channel {
	channel := NewChannel(name, c)
	c.channels[name] = channel
	return channel
}

// Get channel by channel name
func (c *ChannelServive) GetChannel(name string) (*Channel, bool) {
	channel, exists := c.channels[name]
	return channel, exists
}

// Get all members in channel by channel name
func (c *ChannelServive) GetMembers(name string) []int {
	if channel, ok := c.channels[name]; ok {
		return channel.GetMembers()
	}
	return make([]int, 0)
}

// Destroy channel by channel name
func (c *ChannelServive) DestroyChannel(name string) {
	if channel, ok := c.channels[name]; ok {
		channel.Destroy()
	}
}
