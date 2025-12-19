package main

import (
	"context"
	"fmt"
	"log"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
)

//    ReplicationHandler	Control Data
// 1. 5980  5981    5982
// 2. 5990	5991    5992

func main() {
	for {
		sources := []string{":5971", ":5981", ":5991"}
		targets := []string{":5970", ":5980", ":5990"}

		var src, tgt int
		fmt.Scanf("%d %d", &src, &tgt)

		link(sources[src-1], targets[tgt-1])
	}
}

func link(src, target string) {
	ctx := context.Background()
	client := rpc.NewClient(ctx, src)
	control := controllink.NewControlServiceClient(client)
	_, err := control.SwitchSuccessor(ctx, &controllink.SwitchSuccessorCommand{Address: target})
	if err != nil {
		log.Println(err)
	}
	log.Println("Successor switched")
}
