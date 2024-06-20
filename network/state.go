package network

type State struct {
	State Stage
}

type Stage int

const (
	Idle Stage = iota
	preprepared
	prepared
	committed
)
