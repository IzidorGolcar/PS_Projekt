package main

import (
	"context"
	"log"
	"os"
	"seminarska/internal/common/rpc"
	"seminarska/internal/control/dataplane"
	"seminarska/proto/controllink"
	"seminarska/proto/razpravljalnica"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockControlService struct {
	rpcServer *rpc.Server
	nodes     []dataplane.NodeDescriptor
	manager   *dataplane.Manager
	handler   *clientHandler
}

func NewMockControlService(dataNodeExec, addr string) MockControlService {
	manager := dataplane.NewManager(dataNodeExec)
	nodes := launchDataNodes(manager)
	h := newClientHandler(nodes)
	return MockControlService{
		handler:   h,
		nodes:     nodes,
		manager:   manager,
		rpcServer: rpc.NewServer(context.Background(), h, addr),
	}
}

func launchDataNodes(manager *dataplane.Manager) []dataplane.NodeDescriptor {
	head, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node1", os.DevNull,
		"secret", ":6971", ":6981", ":6991",
	))
	if err != nil {
		panic(err)
	}
	mid, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node2", os.DevNull,
		"secret", ":6972", ":6982", ":6992",
	))
	if err != nil {
		panic(err)
	}
	tail, err := manager.StartNewDataNode(dataplane.NewNodeConfig(
		"node3", os.DevNull,
		"secret", ":6973", ":6983", ":6993",
	))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second) // Give nodes time to start

	err = manager.SwitchNodeRole(head, controllink.NodeRole_MessageReader)
	if err != nil {
		panic(err)
	}
	err = manager.SwitchNodeRole(mid, controllink.NodeRole_Relay)
	if err != nil {
		panic(err)
	}
	err = manager.SwitchNodeRole(tail, controllink.NodeRole_MessageConfirmer)
	if err != nil {
		panic(err)
	}

	err = manager.SwitchDataNodeSuccessor(head, mid)
	if err != nil {
		panic(err)
	}
	err = manager.SwitchDataNodeSuccessor(mid, tail)
	if err != nil {
		panic(err)
	}

	return []dataplane.NodeDescriptor{*head, *mid, *tail}

}

func (s *MockControlService) Shutdown() {
	for _, n := range s.nodes {
		err := s.manager.TerminateDataNode(&n)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type clientHandler struct {
	nodes []dataplane.NodeDescriptor
	razpravljalnica.UnimplementedControlPlaneServer
}

func newClientHandler(nodes []dataplane.NodeDescriptor) *clientHandler {
	return &clientHandler{nodes: nodes}
}

func (h *clientHandler) Register(grpcServer *grpc.Server) {
	razpravljalnica.RegisterControlPlaneServer(grpcServer, h)
}

func (h *clientHandler) GetClusterState(
	_ context.Context, _ *emptypb.Empty,
) (*razpravljalnica.GetClusterStateResponse, error) {
	return &razpravljalnica.GetClusterStateResponse{
		Head: h.nodes[0].NodeInfo(),
		Tail: h.nodes[2].NodeInfo(),
	}, nil
}
func (h *clientHandler) GetSubcscriptionNode(
	_ context.Context, _ *razpravljalnica.SubscriptionNodeRequest,
) (*razpravljalnica.SubscriptionNodeResponse, error) {
	return &razpravljalnica.SubscriptionNodeResponse{
		SubscribeToken: h.nodes[1].SubscriptionToken(),
		Node:           h.nodes[1].NodeInfo(),
	}, nil
}
