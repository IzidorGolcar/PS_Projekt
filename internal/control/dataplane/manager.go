package dataplane

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
	"seminarska/proto/razpravljalnica"
	"syscall"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Manager struct {
	dataExecPath string
	rpcClient    *rpc.Client
}

// NewManager creates a new client for controlling the data plane.
// dataExecPath is the path to the data node executable.
func NewManager(dataExecPath string) *Manager {
	return &Manager{
		dataExecPath: dataExecPath,
	}
}

type NodeConfig struct {
	Id                    string
	LoggerPath            string
	SubscriptionToken     string
	ControlAddress        string
	DataChainAddresses    string
	ClientRequestsAddress string
}

type NodeDescriptor struct {
	pid    int
	role   controllink.NodeRole
	config NodeConfig
}

func (n NodeDescriptor) Config() NodeConfig {
	return n.config
}

func (n NodeDescriptor) Role() controllink.NodeRole {
	return n.role
}

func (n NodeDescriptor) SubscriptionToken() string {
	return n.config.SubscriptionToken
}

func (n NodeDescriptor) NodeInfo() *razpravljalnica.NodeInfo {
	return &razpravljalnica.NodeInfo{
		NodeId:  n.config.Id,
		Address: n.config.ClientRequestsAddress,
	}
}

func NewNodeConfig(
	id string,
	loggerPath string,
	subscriptionToken string,
	controlAddress string,
	dataChainAddresses string,
	clientRequestsAddress string,
) NodeConfig {
	return NodeConfig{
		Id: id, LoggerPath: loggerPath,
		ControlAddress:        controlAddress,
		DataChainAddresses:    dataChainAddresses,
		ClientRequestsAddress: clientRequestsAddress,
		SubscriptionToken:     subscriptionToken,
	}
}

func (c *Manager) StartNewDataNode(cfg NodeConfig) (*NodeDescriptor, error) {
	cmd := exec.Command(
		c.dataExecPath,
		"-id", cfg.Id, "-o", cfg.LoggerPath, "-control",
		cfg.ControlAddress, "-chain", cfg.DataChainAddresses,
		"-service", cfg.ClientRequestsAddress, "-token", cfg.SubscriptionToken,
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	descriptor := &NodeDescriptor{
		config: cfg,
		pid:    cmd.Process.Pid,
		role:   controllink.NodeRole_Relay, // default role
	}

	return descriptor, nil
}

func (c *Manager) Ping(node *NodeDescriptor) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.ControlAddress))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := control.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c *Manager) SwitchNodeRole(node *NodeDescriptor, newRole controllink.NodeRole) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.ControlAddress))
	_, err := control.SwitchRole(context.Background(), &controllink.SwitchRoleCommand{Role: newRole})
	if err != nil {
		node.role = newRole
	}
	return err
}

func (c *Manager) SwitchDataNodeSuccessor(node *NodeDescriptor, successor *NodeDescriptor) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.ControlAddress))
	addr := ""
	if successor != nil {
		addr = successor.config.DataChainAddresses
	}
	_, err := control.SwitchSuccessor(context.Background(), &controllink.SwitchSuccessorCommand{Address: addr})
	return err
}

func (c *Manager) DisconnectDataNodeSuccessor(node *NodeDescriptor) error {
	return c.SwitchDataNodeSuccessor(node, nil)
}

func (c *Manager) TerminateDataNode(node *NodeDescriptor) error {
	proc, err := os.FindProcess(node.pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	err = proc.Signal(syscall.SIGINT)
	if err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}
	return nil
}
