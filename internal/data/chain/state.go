package chain

import (
	"fmt"
	"sync"
)

type NodeState struct {
	Position
	Role
	illegal bool
}

func NewNodeState(position Position, role Role) NodeState {
	return NodeState{Position: position, Role: role}
}

func (s NodeState) String() string {
	if s.illegal {
		return "(Illegal)"
	}
	return fmt.Sprintf("(%s; %s)", s.Position, s.Role)
}

//go:generate stringer -type=Position
type Position int

const (
	Head Position = iota
	Middle
	Tail
	Single
)

//go:generate stringer -type=Role
type Role int

const (
	Relay Role = iota
	Reader
	Confirmer
	ReaderConfirmer
)

//go:generate stringer -type=event
type event int

const (
	PredecessorConnect event = iota
	SuccessorConnect
	PredecessorDisconnect
	SuccessorDisconnect
	RoleRelay
	RoleReader
	RoleConfirmer
	RoleReaderConfirmer
)

type NodeDFA struct {
	mx        *sync.Mutex
	states    chan NodeState
	lastState NodeState
}

func NewNodeDFA() *NodeDFA {
	s := &NodeDFA{
		states:    make(chan NodeState, 1),
		lastState: NewNodeState(Single, ReaderConfirmer),
	}
	return s
}

func (d *NodeDFA) States() <-chan NodeState {
	return d.states
}

func (d *NodeDFA) Emit(e event) error {
	switch e {
	case PredecessorConnect:
		if d.lastState.Role == ReaderConfirmer ||
			d.lastState.Role == Reader {
			return illegalTransitionError(d.lastState, e)
		}
		switch d.lastState.Position {
		case Single:
			d.lastState.Position = Tail
		case Head:
			d.lastState.Position = Middle
		default:
			return illegalTransitionError(d.lastState, e)
		}
	case SuccessorConnect:
		if d.lastState.Role == ReaderConfirmer ||
			d.lastState.Role == Confirmer {
			return illegalTransitionError(d.lastState, e)
		}
		switch d.lastState.Position {
		case Single:
			d.lastState.Position = Head
		case Tail:
			d.lastState.Position = Middle
		default:
			return illegalTransitionError(d.lastState, e)
		}
	case PredecessorDisconnect:
		switch d.lastState.Position {
		case Tail:
			d.lastState.Position = Single
		case Middle:
			d.lastState.Position = Head
		default:
			return illegalTransitionError(d.lastState, e)
		}
	case SuccessorDisconnect:
		switch d.lastState.Position {
		case Head:
			d.lastState.Position = Single
		case Middle:
			d.lastState.Position = Tail
		default:
			return illegalTransitionError(d.lastState, e)
		}
	case RoleConfirmer:
		if d.lastState.Position == Single ||
			d.lastState.Position == Tail {
			d.lastState.Role = Confirmer
		} else {
			return illegalTransitionError(d.lastState, e)
		}
	case RoleReader:
		if d.lastState.Position == Single ||
			d.lastState.Position == Head {
			d.lastState.Role = Reader
		} else {
			return illegalTransitionError(d.lastState, e)
		}
	case RoleReaderConfirmer:
		if d.lastState.Position == Single {
			d.lastState.Role = ReaderConfirmer
		} else {
			return illegalTransitionError(d.lastState, e)
		}
	case RoleRelay:
		d.lastState.Role = Relay
	}
	d.states <- d.lastState
	return nil
}

func illegalTransitionError(state NodeState, t event) error {
	return fmt.Errorf("illegal transition: %s%s", t, state)
}
