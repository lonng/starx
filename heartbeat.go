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
		for _, session := range sessionService.sessionAddrMaps {
			if session.status == SS_WORKING {
				session.RawConn.Write(pack(PACKET_HEARTBEAT, nil))
				session.heartbeat()
			}
		}
	}
}
