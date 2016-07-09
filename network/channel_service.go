package network

import (
	"sync"
)

var ChannelServive = newChannelServive()

type channelServive struct {
	channels map[string]*Channel // all server channels
	sync.RWMutex
}

func newChannelServive() *channelServive {
	return &channelServive{
		channels: make(map[string]*Channel),
	}
}

func (c *channelServive) NewChannel(name string) *Channel {
	channel := newChannel(name, c)
	c.Lock()
	defer c.Unlock()
	c.channels[name] = channel
	return channel
}

// Get channel by channel name
func (c *channelServive) Channel(name string) (*Channel, bool) {
	c.RLock()
	defer c.RUnlock()
	return c.channels[name]
}

// Get all members in channel by channel name
func (c *channelServive) Members(name string) []uint64 {
	c.RLock()
	defer c.RUnlock()
	if channel, ok := c.channels[name]; ok {
		return channel.Members()
	}
	return make([]uint64, 0)
}

// Destroy channel by channel name
func (c *channelServive) DestroyChannel(name string) {
	c.RLock()
	c.RUnlock()
	if channel, ok := c.channels[name]; ok {
		channel.Destroy()
	}
}
