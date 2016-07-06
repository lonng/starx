package starx

import (
	"time"
)

type Timer struct {
	ticker     *time.Ticker
	end        chan bool
	limitCount int
	counter    int
}

func (t *Timer) Stop() {
	t.end <- true
}

func Register(d time.Duration, fn func()) *Timer {
	t := &Timer{
		ticker: time.NewTicker(d),
		end:    make(chan bool, 1),
	}
	go func() {
	loop:
		for {
			select {
			case <-t.ticker.C:
				fn()
			case <-t.end:
				t.ticker.Stop()
				break loop
			}
		}
	}()
	return t
}

func RegisterCount(d time.Duration, fn func(), count int) *Timer {
	t := &Timer{
		ticker:     time.NewTicker(d),
		end:        make(chan bool, 1),
		limitCount: count,
		counter:    0,
	}
	go func() {
	loop:
		for {
			select {
			case <-t.ticker.C:
				t.counter++
				if t.counter > t.limitCount {
					t.ticker.Stop()
					break loop
				}
				fn()
			case <-t.end:
				t.ticker.Stop()
				break loop
			}
		}
	}()
	return t
}
