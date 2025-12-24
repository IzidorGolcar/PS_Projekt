package main

import (
	"context"
	"fmt"
	"log"

	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
)

// **************************************
//          MOCK CONTROL SERVICE
// **************************************

func main() {
	sources := []string{":5971", ":5981", ":5991"}
	targets := []string{":5970", ":5980", ":5990"}

	for {
		var cmd string
		fmt.Scan(&cmd)

		switch cmd {
		case "connect":
			var x, y int
			fmt.Scan(&x, &y)
			connect(sources[x-1], targets[y-1])

		case "role":
			var x int
			var r string
			fmt.Scan(&x, &r)
			setRole(sources[x-1], r)

		default:
			log.Println("unknown command:", cmd)
		}
	}
}

func connect(src, target string) {
	ctx := context.Background()
	client := rpc.NewClient(ctx, src)
	control := controllink.NewControlServiceClient(client)

	_, err := control.SwitchSuccessor(
		ctx,
		&controllink.SwitchSuccessorCommand{Address: target},
	)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Successor switched")
}

func setRole(src, role string) {
	ctx := context.Background()
	client := rpc.NewClient(ctx, src)
	control := controllink.NewControlServiceClient(client)

	var roleEnum controllink.NodeRole
	switch role {
	case "rl":
		roleEnum = controllink.NodeRole_Relay
	case "r":
		roleEnum = controllink.NodeRole_MessageReader
	case "c":
		roleEnum = controllink.NodeRole_MessageConfirmer
	case "rc":
		roleEnum = controllink.NodeRole_MessageReaderConfirmer
	}

	_, err := control.SwitchRole(ctx, &controllink.SwitchRoleCommand{Role: roleEnum})
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Role switched to", role)
}
