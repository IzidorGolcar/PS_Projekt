package link

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, _ []string) {
	url := fmt.Sprintf("http://%s/join?id=%s&addr=%s", srcNodeAddr, targetNodeId, targetNodeAddr)
	res, err := http.Get(url)
	if err != nil {
		cmd.PrintErrln(err)
		return
	}
	defer res.Body.Close()
}
