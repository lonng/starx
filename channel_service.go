package starx

type ChannelServive struct {
	channels map[string]*Channel // all server channels
}

func newChannelServive() *ChannelServive {
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
