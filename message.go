package starx

import (
	"fmt"
)

type Message struct {
	Route string
	Body  []byte
}

func (this *Message) String() string {
	return fmt.Sprintf("Route: %s, Body: %s",
		this.Route,
		this.Body)
}
