package proto

//go:generate protoc --experimental_allow_proto3_optional --go_out=.. --go-grpc_out=.. --go_opt=module=seminarska --go-grpc_opt=module=seminarska razpravljalnica.proto datalink.proto control.proto raft.proto
