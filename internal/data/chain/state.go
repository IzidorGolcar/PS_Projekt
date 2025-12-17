package chain

import "context"

//go:generate stringer -type=NodeState
type NodeState int

const (
	Head NodeState = iota
	Middle
	Tail
	SingleNode
	IllegalState
)

type event int

const (
	predecessorConnect event = iota
	successorConnect
	predecessorDisconnect
	successorDisconnect
)

type nodeDFA struct {
	ctx       context.Context
	events    chan event
	states    chan NodeState
	lastState NodeState
}

func newNodeDFA(ctx context.Context) *nodeDFA {
	s := &nodeDFA{
		events:    make(chan event, 1),
		states:    make(chan NodeState, 1),
		ctx:       ctx,
		lastState: SingleNode,
	}
	go s.run()
	return s
}

func (d *nodeDFA) run() {
	d.states <- d.lastState
	for {
		select {
		case <-d.ctx.Done():
			return
		case t := <-d.events:
			d.onEvent(t)
		}
	}
}

func (d *nodeDFA) state() <-chan NodeState {
	return d.states
}

func (d *nodeDFA) emit(e event) {
	d.events <- e
}

func (d *nodeDFA) onEvent(t event) {
	switch t {
	case predecessorConnect:
		switch d.lastState {
		case SingleNode:
			d.lastState = Tail
		case Head:
			d.lastState = Middle
		default:
			d.lastState = IllegalState
		}
	case successorConnect:
		switch d.lastState {
		case SingleNode:
			d.lastState = Head
		case Tail:
			d.lastState = Middle
		default:
			d.lastState = IllegalState
		}
	case predecessorDisconnect:
		switch d.lastState {
		case Tail:
			d.lastState = SingleNode
		case Middle:
			d.lastState = Head
		default:
			d.lastState = IllegalState
		}
	case successorDisconnect:
		switch d.lastState {
		case Head:
			d.lastState = SingleNode
		case Middle:
			d.lastState = Tail
		default:
			d.lastState = IllegalState
		}
	}
	d.states <- d.lastState
}
