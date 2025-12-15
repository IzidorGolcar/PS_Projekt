package chain

import (
	"context"
	"errors"
	"time"
)

type connectedNode struct {
	*Client
	cancel context.CancelFunc
	ctx    context.Context
}

func (n *connectedNode) disconnect(timeout time.Duration) error {
	disconnectCtx, cancel := context.WithTimeout(n.ctx, timeout)
	defer cancel()
	n.cancel()
	select {
	case <-n.Done():
		return nil
	case <-disconnectCtx.Done():
		return errors.New("failed to disconnect node")
	}
}

func newConnectedNode(ctx context.Context, address string) *connectedNode {
	clientCtx, cancel := context.WithCancel(ctx)
	return &connectedNode{
		Client: NewClient(clientCtx, address),
		ctx:    clientCtx,
		cancel: cancel,
	}
}
