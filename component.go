package starx

type Component interface {
	Setup()
}

type HandlerComponent interface {
	Setup()
}

type RpcComponent interface {
	Setup()
}
