package starx

type Component interface {
	Init()
	AfterInit()
	BeforeShutdown()
	Shutdown()
}

type NopComponent struct{}

func (c *NopComponent) Init()           {}
func (c *NopComponent) AfterInit()      {}
func (c *NopComponent) BeforeShutdown() {}
func (c *NopComponent) Shutdown()       {}
