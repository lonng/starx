package starx

type Component interface {
	Init()
	AfterInit()
	BeforeShutdown()
	Shutdown()
}
