package network

type Component interface {
	Init()
	AfterInit()
	BeforeShutdown()
	Shutdown()
}