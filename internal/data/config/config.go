package config

type NodeConfig struct {
	NodeId               string
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

func Load() NodeConfig {
	return *DefaultNodeConfig()
}
