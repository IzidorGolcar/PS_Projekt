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

type NodeManager struct {
	dataExecPath string
	rpcClient    *rpc.Client
}

// NewNodeManager creates a new client for controlling the data plane.
// dataExecPath is the path to the data node executable.
func NewNodeManager(dataExecPath string) *NodeManager {
	return &NodeManager{
		dataExecPath: dataExecPath,
	}
}

type NodeConfig struct {
	Id                    string `json:"id,omitempty"`
	LoggerPath            string `json:"loggerPath,omitempty"`
	SubscriptionToken     string `json:"subscriptionToken,omitempty"`
	ControlAddress        string `json:"controlAddress,omitempty"`
	DataChainAddresses    string `json:"dataChainAddresses,omitempty"`
	ClientRequestsAddress string `json:"clientRequestsAddress,omitempty"`
}

type NodeDescriptor struct {
	Pid    int                  `json:"pid,omitempty"`
	Role   controllink.NodeRole `json:"role,omitempty"`
	Config NodeConfig           `json:"config"`
}

func (n NodeDescriptor) SubscriptionToken() string {
	return n.Config.SubscriptionToken
}

func (n NodeDescriptor) NodeInfo() *razpravljalnica.NodeInfo {
	return &razpravljalnica.NodeInfo{
		NodeId:  n.Config.Id,
		Address: n.Config.ClientRequestsAddress,
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

func (c *NodeManager) StartNewDataNode(cfg NodeConfig) (*NodeDescriptor, error) {
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
		Config: cfg,
		Pid:    cmd.Process.Pid,
		Role:   controllink.NodeRole_Relay, // default role
	}

	return descriptor, nil
}

func (c *NodeManager) Ping(node *NodeDescriptor) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.Config.ControlAddress))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := control.Ping(ctx, &emptypb.Empty{})
	return err
}

func (c *NodeManager) SwitchNodeRole(node *NodeDescriptor, newRole controllink.NodeRole) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.Config.ControlAddress))
	_, err := control.SwitchRole(context.Background(), &controllink.SwitchRoleCommand{Role: newRole})
	if err != nil {
		node.Role = newRole
	}
	return err
}

func (c *NodeManager) SwitchDataNodeSuccessor(node *NodeDescriptor, successor *NodeDescriptor) error {
	control := controllink.NewControlServiceClient(rpc.NewClient(context.Background(), node.Config.ControlAddress))
	addr := ""
	if successor != nil {
		addr = successor.Config.DataChainAddresses
	}
	_, err := control.SwitchSuccessor(context.Background(), &controllink.SwitchSuccessorCommand{Address: addr})
	return err
}

func (c *NodeManager) DisconnectDataNodeSuccessor(node *NodeDescriptor) error {
	return c.SwitchDataNodeSuccessor(node, nil)
}

func (c *NodeManager) TerminateDataNode(node *NodeDescriptor) error {
	proc, err := os.FindProcess(node.Pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}
	err = proc.Signal(syscall.SIGINT)
	if err != nil {
		return fmt.Errorf("failed to terminate process: %w", err)
	}
	return nil
}
