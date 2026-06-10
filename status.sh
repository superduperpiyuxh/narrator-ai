#!/bin/bash
# Check status of NarratorAI backend

echo "=== NarratorAI Status ==="

# Check if process exists
if [ -f /tmp/narrator-ai.pid ]; then
    PID=$(cat /tmp/narrator-ai.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "Process: RUNNING (PID $PID)"
    else
        echo "Process: DEAD (stale PID file)"
    fi
else
    echo "Process: NOT STARTED (no PID file)"
fi

# Check health endpoint
HEALTH=$(curl -s http://localhost:8080/health 2>/dev/null)
if [ -n "$HEALTH" ]; then
    echo "Health:  $HEALTH"
else
    echo "Health:  UNREACHABLE"
fi

# Check event count
STATS=$(curl -s http://localhost:8080/api/stats 2>/dev/null)
if [ -n "$STATS" ]; then
    TOTAL=$(echo "$STATS" | grep -o '"total_events":[0-9]*' | cut -d: -f2)
    echo "Events:  ${TOTAL:-0}"
else
    echo "Events:  UNKNOWN"
fi
