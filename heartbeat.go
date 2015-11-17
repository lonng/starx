package starx

import (
	"time"
)

type HeartbeatService struct {
	ticker *time.Ticker
}

func NewHeartbeatService() *HeartbeatService {
	return &HeartbeatService{ticker: time.NewTicker(heartbeatInternal)}
}

func (h *HeartbeatService) start() {
	for {
		<-h.ticker.C
		Net.heartbeat()
	}
}
