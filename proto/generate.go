package proto

//go:generate protoc --proto_path=. --go_out=paths=import:.. --go-grpc_out=paths=import:.. razpravljalnica.proto datalink.proto control.proto raft.proto
//go:generate sh -c "mv ../seminarska/proto/razpravljalnica/* razpravljalnica/ && mv ../seminarska/proto/datalink/* datalink/ && mv ../seminarska/proto/controllink/* controllink/ && mv ../seminarska/proto/raft/* raft/ && rm -rf ../seminarska/"
