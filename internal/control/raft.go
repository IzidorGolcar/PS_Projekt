package control

import (
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
)

func SetupRaft(
	nodeID string,
	raftAddr string,
	dataDir string,
	fsm raft.FSM,
	bootstrap bool,
) (*raft.Raft, error) {

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(
		filepath.Join(dataDir, "raft.db"),
	)
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(dataDir, 1, os.Stdout)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return nil, err
	}

	transport, err := raft.NewTCPTransport(
		raftAddr,
		addr,
		3,
		10*time.Second,
		os.Stdout,
	)
	if err != nil {
		return nil, err
	}

	r, err := raft.NewRaft(
		config,
		fsm,
		logStore,
		logStore,
		snapshots,
		transport,
	)
	if err != nil {
		return nil, err
	}

	if bootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		r.BootstrapCluster(cfg)
		log.Println("bootstrapped cluster")
	}

	return r, nil
}
