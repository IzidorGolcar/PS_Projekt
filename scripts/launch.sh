go run cmd/control/main.go launch \
  --node-id node1 \
  --raft-addr 127.0.0.1:5301 \
  --http-addr :8301 \
  --rpc-addr :8080 \
  --data-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service" \
  --log-dir "/Users/izidor/Downloads/dataplane.log" \
  --bootstrap

go run cmd/control/main.go launch \
  --node-id node2 \
  --raft-addr 127.0.0.1:5302 \
  --http-addr :8302 \
  --rpc-addr :8081 \
  --log-dir "/Users/izidor/Downloads/dataplane.log" \
  --data-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service"

go run cmd/control/main.go launch \
  --node-id node3 \
  --raft-addr 127.0.0.1:5303 \
  --http-addr :8303 \
  --rpc-addr :8082 \
  --log-dir "/Users/izidor/Downloads/dataplane.log" \
  --data-exec "/Users/izidor/Code/UNI/PS/seminarska/build/data_service"


go run cmd/control/main.go link \
  --src "localhost:8301" \
  --target "127.0.0.1:5302" \
  --target-id node2

go run cmd/control/main.go link \
  --src "localhost:8301" \
  --target "127.0.0.1:5303" \
  --target-id node3

go run cmd/control/main.go state \
  --addr "localhost:8301"

