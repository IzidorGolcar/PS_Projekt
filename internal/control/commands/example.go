package commands

import (
	"fmt"
	"log"
	"seminarska/proto/controllink"
	"time"
)

func Example() {

	// 1. Start a new node
	fmt.Println("Starting a new node")
	c := NewClient("/Users/izidor/Code/UNI/PS/seminarska/build/data_service")
	node, err := c.StartNewDataNode(
		NewNodeConfig(
			"node1", "/Users/izidor/Downloads/node1.log",
			"secret!!!", ":6971", ":6981", ":6991",
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)

	// 2. Change role
	time.Sleep(time.Second * 3)
	fmt.Println("Changing role")
	err = c.SwitchNodeRole(node, controllink.NodeRole_MessageReaderConfirmer)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)

	// 3. Ping node
	time.Sleep(time.Second * 20)
	fmt.Println("Pinging node")
	err = c.Ping(node)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Terminate node
	time.Sleep(time.Second * 3)
	fmt.Println("Terminating node")
	err = c.TerminateDataNode(node)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(node)
	fmt.Println("Done")
}
