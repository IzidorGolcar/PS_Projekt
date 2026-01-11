# node 1
go run cmd/controlv2/main.go \
  --id node1 \
  --raft 127.0.0.1:5301 \
  --http :8301 \
  --rpc :8080 \
  --data data1 \
  --node-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service" \
  --bootstrap

# node 2
go run cmd/controlv2/main.go \
  --id node2 \
  --raft 127.0.0.1:5302 \
  --http :8302 \
  --rpc :8081 \
  --node-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service" \
  --data data2

# node 3
go run cmd/controlv2/main.go \
  --id node3 \
  --raft 127.0.0.1:5303 \
  --http :8303 \
  --rpc :8082 \
  --node-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service" \
  --data data3

# join node2
curl "localhost:8301/join?id=node2&addr=127.0.0.1:5302"
curl "localhost:8301/join?id=node3&addr=127.0.0.1:5303"
