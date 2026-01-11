#!/bin/bash

# Test script for Raft consensus with 3 control plane nodes
# Usage: ./scripts/test_raft.sh [start|stop|status|logs]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
LOG_DIR="$PROJECT_DIR/logs"
PID_DIR="$PROJECT_DIR/.pids"

# Port configuration - change these variables to run on different ports
CLIENT_PORT_NODE1=8080
CLIENT_PORT_NODE2=8081
CLIENT_PORT_NODE3=8082

RAFT_PORT_NODE1=7080
RAFT_PORT_NODE2=7081
RAFT_PORT_NODE3=7082

# Create directories
mkdir -p "$LOG_DIR" "$PID_DIR"

# Build the control plane binary
build() {
    echo "Building control plane..."
    cd "$PROJECT_DIR"
    go build -o bin/control ./cmd/control
    echo "Build complete: bin/control"
}

# Start all 3 Raft nodes
start() {
    "$PROJECT_DIR/scripts/build.sh"
    
    echo "Starting Raft cluster (3 nodes)..."
    
    # Node 1
    echo "Starting node1 (client: :$CLIENT_PORT_NODE1, raft: :$RAFT_PORT_NODE1)..."
    "$PROJECT_DIR/build/control_service" \
        -id node1 \
        -addr :$CLIENT_PORT_NODE1 \
        -raft-addr :$RAFT_PORT_NODE1 \
        -peers :$RAFT_PORT_NODE2,:$RAFT_PORT_NODE3 \
        -peer-ids node2,node3 \
        -data-exec "$PROJECT_DIR/build/data_service" \
        > "$LOG_DIR/node1.log" 2>&1 &
    echo $! > "$PID_DIR/node1.pid"
    
    # Node 2
    echo "Starting node2 (client: :$CLIENT_PORT_NODE2, raft: :$RAFT_PORT_NODE2)..."
    "$PROJECT_DIR/build/control_service" \
        -id node2 \
        -addr :$CLIENT_PORT_NODE2 \
        -raft-addr :$RAFT_PORT_NODE2 \
        -peers :$RAFT_PORT_NODE1,:$RAFT_PORT_NODE3 \
        -peer-ids node1,node3 \
        -data-exec "$PROJECT_DIR/build/data_service" \
        > "$LOG_DIR/node2.log" 2>&1 &
    echo $! > "$PID_DIR/node2.pid"
    
    # Node 3
    echo "Starting node3 (client: :$CLIENT_PORT_NODE3, raft: :$RAFT_PORT_NODE3)..."
    "$PROJECT_DIR/build/control_service" \
        -id node3 \
        -addr :$CLIENT_PORT_NODE3 \
        -raft-addr :$RAFT_PORT_NODE3 \
        -peers :$RAFT_PORT_NODE1,:$RAFT_PORT_NODE2 \
        -peer-ids node1,node2 \
        -data-exec "$PROJECT_DIR/build/data_service" \
        > "$LOG_DIR/node3.log" 2>&1 &
    echo $! > "$PID_DIR/node3.pid"
    
    echo ""
    echo "Raft cluster started!"
    echo "  Node1: client=:$CLIENT_PORT_NODE1, raft=:$RAFT_PORT_NODE1 (log: $LOG_DIR/node1.log)"
    echo "  Node2: client=:$CLIENT_PORT_NODE2, raft=:$RAFT_PORT_NODE2 (log: $LOG_DIR/node2.log)"
    echo "  Node3: client=:$CLIENT_PORT_NODE3, raft=:$RAFT_PORT_NODE3 (log: $LOG_DIR/node3.log)"
    echo ""
    echo "Use './scripts/test_raft.sh logs' to view logs"
    echo "Use './scripts/test_raft.sh stop' to stop all nodes"
}

# Stop all nodes
stop() {
    echo "Stopping Raft cluster..."
    
    for node in node1 node2 node3; do
        if [ -f "$PID_DIR/$node.pid" ]; then
            pid=$(cat "$PID_DIR/$node.pid")
            if kill -0 "$pid" 2>/dev/null; then
                echo "Stopping $node (PID: $pid)..."
                kill -INT "$pid" 2>/dev/null || true
            fi
            rm -f "$PID_DIR/$node.pid"
        fi
    done
    
    echo "All nodes stopped."
}

# Check status of nodes
status() {
    echo "Raft cluster status:"
    echo ""
    
    for node in node1 node2 node3; do
        if [ -f "$PID_DIR/$node.pid" ]; then
            pid=$(cat "$PID_DIR/$node.pid")
            if kill -0 "$pid" 2>/dev/null; then
                echo "  $node: RUNNING (PID: $pid)"
            else
                echo "  $node: STOPPED (stale PID file)"
                rm -f "$PID_DIR/$node.pid"
            fi
        else
            echo "  $node: STOPPED"
        fi
    done
}

# View logs (tail all 3 logs)
logs() {
    echo "Tailing logs (Ctrl+C to exit)..."
    echo ""
    tail -f "$LOG_DIR/node1.log" "$LOG_DIR/node2.log" "$LOG_DIR/node3.log"
}

# Kill a specific node (for testing failover)
kill_node() {
    node=$1
    if [ -z "$node" ]; then
        echo "Usage: $0 kill <node1|node2|node3>"
        exit 1
    fi
    
    if [ -f "$PID_DIR/$node.pid" ]; then
        pid=$(cat "$PID_DIR/$node.pid")
        echo "Killing $node (PID: $pid)..."
        kill -9 "$pid" 2>/dev/null || true
        rm -f "$PID_DIR/$node.pid"
        echo "$node killed. Watch the logs to see leader election!"
    else
        echo "$node is not running."
    fi
}

# Main
case "${1:-}" in
    start)  start ;;
    stop)   stop ;;
    status) status ;;
    logs)   logs ;;
    kill)   kill_node "$2" ;;
    *)
        echo "Usage: $0 {start|stop|status|logs|kill <node>}"
        echo ""
        echo "Commands:"
        echo "  start  - Build and start 3-node Raft cluster"
        echo "  stop   - Stop all nodes gracefully"
        echo "  status - Check which nodes are running"
        echo "  logs   - Tail logs from all nodes"
        echo "  kill   - Kill a specific node (e.g., 'kill node1')"
        exit 1
        ;;
esac

