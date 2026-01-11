package control

import (
	"encoding/json"
	"log"
	"net/http"
	"seminarska/internal/control/dataplane"

	"github.com/hashicorp/raft"
)

type NodeStateReport struct {
	State    string                      `json:"state"`
	Snapshot []*dataplane.NodeDescriptor `json:"snapshot"`
}

func StartHTTP(addr string, r *raft.Raft, fms *ChainFSM) {

	http.HandleFunc("/join", func(w http.ResponseWriter, req *http.Request) {
		if r.State() != raft.Leader {
			http.Error(w, "not leader", 403)
			return
		}

		id := req.URL.Query().Get("id")
		addr := req.URL.Query().Get("addr")

		f := r.AddVoter(
			raft.ServerID(id),
			raft.ServerAddress(addr),
			0,
			0,
		)

		if err := f.Error(); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, req *http.Request) {
		s := NodeStateReport{
			State:    r.State().String(),
			Snapshot: fms.nodes,
		}
		_ = json.NewEncoder(w).Encode(s)
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}
