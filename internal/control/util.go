package control

import "fmt"

const (
	clientPortOffset  = 10_080
	controlPortOffset = 20_080
	dataPortOffset    = 30_080
)

func getNodePorts(nodeId int) (string, string, string) {
	return fmt.Sprintf(":%d", clientPortOffset+nodeId),
		fmt.Sprintf(":%d", controlPortOffset+nodeId),
		fmt.Sprintf(":%d", dataPortOffset+nodeId)
}
