package config

type NodeConfig struct {
	ServiceAddress       string
	ChainListenerAddress string
	ChainTargetAddress   string
	TailAddress          string
}

func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		ServiceAddress:       ":0",
		ChainListenerAddress: ":0",
		ChainTargetAddress:   ":0",
		TailAddress:          ":0",
	}
}
