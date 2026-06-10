#!/bin/bash
# Start the NarratorAI backend server
cd "$(dirname "$0")/backend"

# Kill any existing instance
pkill -9 -f "narrator-ai" 2>/dev/null
sleep 1

# Set environment
export DATA_DIR="$(dirname "$0")/data/sample_json_20260301"
export PORT=8080

# Start in background
nohup ./narrator-ai > /tmp/narrator-ai.log 2>&1 &
echo $! > /tmp/narrator-ai.pid

sleep 2

# Verify it's running
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "Server started successfully on port $PORT"
    echo "PID: $(cat /tmp/narrator-ai.pid)"
    echo "Health: $(curl -s http://localhost:8080/health)"
else
    echo "FAILED to start. Check /tmp/narrator-ai.log"
    exit 1
fi
