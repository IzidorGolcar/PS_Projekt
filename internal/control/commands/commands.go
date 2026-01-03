package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
	"syscall"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	dataExecPath string
	rpcClient    *rpc.Client
}

// NewClient creates a new client for controlling the data plane.
// dataExecPath is the path to the data node executable.
func NewClient(dataExecPath string) *Client {
	return &Client{
		dataExecPath: dataExecPath,
	}
}

type NodeConfig struct {
	id                    string
	loggerPath            string
	subscriptionToken     string
	controlAddress        string
	dataChainAddresses    string
	clientRequestsAddress string
}

type NodeDescriptor struct {
	pid    int
	role   controllink.NodeRole
	config NodeConfig
}

func (n NodeDescriptor) Role() controllink.NodeRole {
	return n.role
}

func (n NodeDescriptor) SubscriptionToken() string {
	return n.config.subscriptionToken
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
		id: id, loggerPath: loggerPath,
		controlAddress:        controlAddress,
		dataChainAddresses:    dataChainAddresses,
		clientRequestsAddress: clientRequestsAddress,
		subscriptionToken:     subscriptionToken,
	}
}

func (c *Client) StartNewDataNode(cfg NodeConfig) (*NodeDescriptor, error) {
	cmd := exec.Command(
		c.dataExecPath,
		"-id", cfg.id, "-o", cfg.loggerPath, "-control",
		cfg.controlAddress, "-chain", cfg.dataChainAddresses,
		"-service", cfg.clientRequestsAddress, "-token", cfg.subscriptionToken,
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

func (c *Client) Ping(node *NodeDescriptor) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.controlAddress))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := control.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c *Client) SwitchNodeRole(node *NodeDescriptor, newRole controllink.NodeRole) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.controlAddress))
	_, err := control.SwitchRole(context.Background(), &controllink.SwitchRoleCommand{Role: newRole})
	if err != nil {
		node.role = newRole
	}
	return err
}

func (c *Client) SwitchDataNodeSuccessor(node *NodeDescriptor, newSuccessorAddr string) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.config.controlAddress))
	_, err := control.SwitchSuccessor(context.Background(), &controllink.SwitchSuccessorCommand{Address: newSuccessorAddr})
	return err
}

func (c *Client) DisconnectDataNodeSuccessor(node *NodeDescriptor) error {
	return c.SwitchDataNodeSuccessor(node, "")
}

func (c *Client) TerminateDataNode(node *NodeDescriptor) error {
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
