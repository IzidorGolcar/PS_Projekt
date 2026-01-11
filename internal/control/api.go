package control

import (
	"context"
	"errors"
	"seminarska/internal/control/dataplane"
	"seminarska/proto/razpravljalnica"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChainState interface {
	Head() *dataplane.NodeDescriptor
	Mid() *dataplane.NodeDescriptor
	Tail() *dataplane.NodeDescriptor
}

type clientHandler struct {
	state ChainState
	razpravljalnica.UnimplementedControlPlaneServer
}

func newClientHandler(state ChainState) *clientHandler {
	return &clientHandler{state: state}
}

func (h *clientHandler) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterControlPlaneServer(grpcServer, h)
}

func (h *clientHandler) GetClusterState(
	_ context.Context, _ *emptypb.Empty,
) (*razpravljalnica.GetClusterStateResponse, error) {
	head := h.state.Head()
	tail := h.state.Tail()
	if head == nil || tail == nil {
		return nil, errors.New("cluster not initialized")
	}
	return &razpravljalnica.GetClusterStateResponse{
		Head: head.NodeInfo(),
		Tail: tail.NodeInfo(),
	}, nil
}
func (h *clientHandler) GetSubcscriptionNode(
	_ context.Context, _ *razpravljalnica.SubscriptionNodeRequest,
) (*razpravljalnica.SubscriptionNodeResponse, error) {
	handlerNode := h.state.Mid()
	if handlerNode == nil {
		return nil, errors.New("cluster not initialized")
	}
	return &razpravljalnica.SubscriptionNodeResponse{
		SubscribeToken: handlerNode.SubscriptionToken(),
		Node:           handlerNode.NodeInfo(),
	}, nil
}
