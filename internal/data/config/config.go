package config

import "flag"

type NodeConfig struct {
	NodeId                 string
	ServiceAddress         string
	ChainListenerAddress   string
	ControlListenerAddress string
	LogPath                string
	Token                  string
}

func Load() NodeConfig {
	id := flag.String("id", "data-node", "Node ID")
	serviceAddress := flag.String("service", ":0", "Service address")
	chainListenerAddress := flag.String("chain", ":0", "ReplicationHandler listener address")
	controlListenerAddress := flag.String("control", ":0", "Control listener address")
	token := flag.String("token", "", "Token")
	logPath := flag.String("o", "", "Log path")
	flag.Parse()

	return NodeConfig{
		NodeId:                 *id,
		ServiceAddress:         *serviceAddress,
		ChainListenerAddress:   *chainListenerAddress,
		ControlListenerAddress: *controlListenerAddress,
		Token:                  *token,
		LogPath:                *logPath,
	}
}
