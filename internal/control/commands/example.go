package commands

import (
	"fmt"
	"log"
	"seminarska/proto/controllink"
	"time"
)

func Example() {

	// 1. Start a new node

	c := NewClient("/Users/izidor/Code/UNI/PS/seminarska/build/data_service")
	node, err := c.StartNewDataNode(
		NewNodeConfig(
			"node1", "/Users/izidor/Downloads/node1.log",
			":6971", ":6981", ":6991",
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)

	// 2. Change role

	err = c.SwitchNodeRole(node, controllink.NodeRole_MessageReaderConfirmer)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)

	// 3. Terminate node

	time.Sleep(time.Second * 3)
	err = c.TerminateDataNode(node)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)
}
