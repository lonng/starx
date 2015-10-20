package mello

import (
	"time"
)

type Timer struct {
	Name    string
	Tickers []*time.Ticker
}

func NewTimer(app *MelloApp) Timer {
	return Timer{Name: "TimerComponent"}
}

func (this *Timer) Register(d time.Duration, fn func()) {
	ticker := time.NewTicker(d)
	this.Tickers = append(this.Tickers, ticker)
	go func() {
		for {
			select {
			case <-ticker.C:
				{
					fn()
				}

			}
		}
	}()
}
