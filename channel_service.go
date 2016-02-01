package starx

type channelServive struct {
	channels map[string]*Channel // all server channels
}

func newChannelServive() *channelServive {
	return &channelServive{}
}

func (c *channelServive) NewChannel(name string) *Channel {
	channel := NewChannel(name, c)
	c.channels[name] = channel
	return channel
}

// Get channel by channel name
func (c *channelServive) GetChannel(name string) (*Channel, bool) {
	channel, exists := c.channels[name]
	return channel, exists
}

// Get all members in channel by channel name
func (c *channelServive) GetMembers(name string) []int {
	if channel, ok := c.channels[name]; ok {
		return channel.GetMembers()
	}
	return make([]int, 0)
}

// Destroy channel by channel name
func (c *channelServive) DestroyChannel(name string) {
	if channel, ok := c.channels[name]; ok {
		channel.Destroy()
	}
}
