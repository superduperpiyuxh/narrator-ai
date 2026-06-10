#!/bin/bash
# Stop the NarratorAI backend server
PID=$(cat /tmp/narrator-ai.pid 2>/dev/null)
if [ -n "$PID" ]; then
    kill $PID 2>/dev/null
    sleep 1
    rm -f /tmp/narrator-ai.pid
    echo "Server stopped (PID: $PID)"
else
    echo "No server running"
fi
