#!/bin/bash
# Start the NarratorAI backend server
cd "$(dirname "$0")/backend"

# Kill any existing instance (non-blocking)
kill $(cat /tmp/narrator-ai.pid 2>/dev/null) 2>/dev/null
sleep 1

# Set environment
export DATA_DIR="$(dirname "$0")/data/sample_json_20260301"
export PORT=8080

# Start in background with nohup
nohup ./narrator-ai > /tmp/narrator-ai.log 2>&1 &
echo $! > /tmp/narrator-ai.pid

# Wait for server to be ready (max 10 seconds)
for i in $(seq 1 10); do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "Server started on port $PORT (PID: $(cat /tmp/narrator-ai.pid))"
        exit 0
    fi
    sleep 1
done

echo "FAILED to start. Check /tmp/narrator-ai.log"
exit 1
