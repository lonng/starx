package starx

import (
	"time"
)

type heartbeatService struct {
	ticker *time.Ticker
}

func newHeartbeatService() *heartbeatService {
	return &heartbeatService{ticker: time.NewTicker(heartbeatInternal)}
}

func (h *heartbeatService) start() {
	Info("enable heartbeat service")
	for {
		<-h.ticker.C
		defaultNetService.heartbeat()
	}
}
