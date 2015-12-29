package starx

import (
	"time"
)

type HeartbeatService struct {
	ticker *time.Ticker
}

func newHeartbeatService() *HeartbeatService {
	return &HeartbeatService{ticker: time.NewTicker(heartbeatInternal)}
}

func (h *HeartbeatService) start() {
	Info("enable heartbeat service")
	for {
		<-h.ticker.C
		netService.heartbeat()
	}
}
