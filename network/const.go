package network

type networkStatus byte

const (
	_ networkStatus = iota
	statusStart
	statusHandshake
	statusWorking
	statusClosed
)
