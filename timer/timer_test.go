package timer

import (
	"testing"
	"time"
)

func TestRegister(t *testing.T) {
	wait := make(chan bool, 1)
	counter := 0
	timer := Register(10*time.Millisecond, func() {
		counter++
	})

	time.AfterFunc(55*time.Millisecond, func() {
		wait <- true
	})

	<-wait
	timer.Stop()
	if counter != 5 {
		t.Fail()
	}
}

func TestRegisterCount(t *testing.T) {
	wait := make(chan bool, 1)
	counter := 0
	timer := RegisterCount(10*time.Millisecond, func() {
		counter++
	}, 5)

	time.AfterFunc(80*time.Millisecond, func() {
		wait <- true
	})

	<-wait
	timer.Stop()
	if counter != 5 {
		t.Fail()
	}
}
