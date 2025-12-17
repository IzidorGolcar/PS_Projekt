package main

import (
	"context"
	"seminarska/internal/common/rpc"
	"seminarska/proto/controllink"
)

//    Chain	Control Data
// 1. 5980  5981    5982
// 2. 5990	5991    5992

func main() {
	link(":5971", ":5980")
	link(":5981", ":5990")
}

func link(src, target string) {
	ctx := context.Background()
	client := rpc.NewClient(ctx, src)
	control := controllink.NewControlServiceClient(client)
	_, err := control.SwitchSuccessor(ctx, &controllink.SwitchSuccessorCommand{Address: target})
	if err != nil {
		panic(err)
	}
}
