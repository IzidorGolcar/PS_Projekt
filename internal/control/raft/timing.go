package raft

import (
	"math/rand"
	"time"
)

const (
	// MinElectionTimeout is the minimum election timeout
	MinElectionTimeout = 150 * time.Millisecond
	// MaxElectionTimeout is the maximum election timeout
	MaxElectionTimeout = 300 * time.Millisecond
	// HeartbeatInterval is the interval between heartbeats sent by the leader
	HeartbeatInterval = 50 * time.Millisecond
	// HealthCheckInterval is the interval between health checks of data nodes
	HealthCheckInterval = 100 * time.Millisecond
	// HealthCheckTimeout is the timeout for health check requests
	HealthCheckTimeout = 50 * time.Millisecond
)

// randomElectionTimeout returns a random election timeout between MinElectionTimeout and MaxElectionTimeout
// This randomization helps prevent split votes in leader elections
func randomElectionTimeout() time.Duration {
	min := int64(MinElectionTimeout)
	max := int64(MaxElectionTimeout)
	return time.Duration(min + rand.Int63n(max-min))
}
