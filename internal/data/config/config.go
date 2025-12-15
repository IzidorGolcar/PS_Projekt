package config

type NodeConfig struct {
	NodeId               string
	ServiceAddress       string
	ChainListenerAddress string
}

func DefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		ServiceAddress:       ":0",
		ChainListenerAddress: ":0",
	}
}

func Load() NodeConfig {
	return *DefaultNodeConfig()
}
