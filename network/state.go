package network

type State struct {
	currentState Stage
}

type Stage int

const (
	Idle Stage = iota
	preprepared
	prepared
	committed
)
