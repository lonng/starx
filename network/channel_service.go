package network

var ChannelServive = newChannelServive()

type channelServive struct {
	channels map[string]*Channel // all server channels
}

func newChannelServive() *channelServive {
	return &channelServive{make(map[string]*Channel)}
}

func (c *channelServive) NewChannel(name string) *Channel {
	channel := newChannel(name, c)
	c.channels[name] = channel
	return channel
}

// Get channel by channel name
func (c *channelServive) Channel(name string) (*Channel, bool) {
	channel, exists := c.channels[name]
	return channel, exists
}

// Get all members in channel by channel name
func (c *channelServive) Members(name string) []uint64 {
	if channel, ok := c.channels[name]; ok {
		return channel.Members()
	}
	return make([]uint64, 0)
}

// Destroy channel by channel name
func (c *channelServive) DestroyChannel(name string) {
	if channel, ok := c.channels[name]; ok {
		channel.Destroy()
	}
}
