package starx

import (
	"sync"
)

var ChannelService = newChannelServive()

type channelService struct {
	channels map[string]*Channel // all server channels
	sync.RWMutex
}

func newChannelServive() *channelService {
	return &channelService{
		channels: make(map[string]*Channel),
	}
}

func (c *channelService) NewChannel(name string) *Channel {
	c.Lock()
	defer c.Unlock()

	channel := newChannel(name, c)
	c.channels[name] = channel
	return channel
}

// Get channel by channel name
func (c *channelService) Channel(name string) (*Channel, bool) {
	c.RLock()
	defer c.RUnlock()

	channel, ok := c.channels[name]
	return channel, ok
}

// Get all members in channel by channel name
func (c *channelService) Members(name string) []int64 {
	c.RLock()
	defer c.RUnlock()

	if channel, ok := c.channels[name]; ok {
		return channel.Members()
	}
	return make([]int64, 0)
}

// Destroy channel by channel name
func (c *channelService) DestroyChannel(name string) {
	c.RLock()
	c.RUnlock()

	if channel, ok := c.channels[name]; ok {
		channel.Destroy()
	}
}
