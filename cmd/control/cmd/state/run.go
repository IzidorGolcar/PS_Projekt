package state

import (
	"encoding/json"
	"fmt"
	"net/http"
	"seminarska/internal/control"

	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, _ []string) {
	url := fmt.Sprintf("http://%s/state", addr)
	res, err := http.Get(url)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	defer res.Body.Close()
	s := control.NodeStateReport{}
	err = json.NewDecoder(res.Body).Decode(&s)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	printState(s)
}

func printState(s control.NodeStateReport) {
	fmt.Printf("Node state: %s\n", s.State)
	fmt.Println("Data nodes:")
	for _, node := range s.Snapshot {
		fmt.Println(node)
	}
}
