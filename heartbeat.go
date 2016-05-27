package starx

import (
	"github.com/chrislonng/starx/log"
	"time"
)

type heartbeatService struct {
	ticker *time.Ticker
}

func newHeartbeatService() *heartbeatService {
	return &heartbeatService{ticker: time.NewTicker(heartbeatInternal)}
}

func (h *heartbeatService) start() {
	log.Info("enable heartbeat service")
	for {
		<-h.ticker.C
		defaultNetService.heartbeat()
	}
}
