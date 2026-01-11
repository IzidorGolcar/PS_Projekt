package _custom_raft_impl

import "errors"

var (
	// ErrNotLeader is returned when a command is submitted to a non-leader node
	ErrNotLeader = errors.New("not the leader")
	// ErrTimeout is returned when an operation times out
	ErrTimeout = errors.New("operation timed out")
	// ErrNodeNotFound is returned when a node is not found
	ErrNodeNotFound = errors.New("node not found")
	// ErrAlreadyExists is returned when trying to add a node that already exists
	ErrAlreadyExists = errors.New("node already exists")
)
