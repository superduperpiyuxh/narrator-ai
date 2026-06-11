#!/bin/bash
# Start the Nexus backend server
cd "$(dirname "$0")/backend"

# Kill any existing instance (non-blocking)
kill $(cat /tmp/narrator-ai.pid 2>/dev/null) 2>/dev/null
sleep 1

# Load .env file if it exists (dotenv)
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Defaults (override in .env)
export PORT="${PORT:-8080}"
export JWT_SECRET="${JWT_SECRET:-narrator-ai-jwt-secret}"
export DATABASE_PATH="${DATABASE_PATH:-./narratorai.db}"
export DATA_DIR="${DATA_DIR:-$(dirname "$0")/data/sample_json_20260301}"

# Start in background with nohup
nohup ./narrator-ai > /tmp/narrator-ai.log 2>&1 &
echo $! > /tmp/narrator-ai.pid

# Wait for server to be ready (max 10 seconds)
for i in $(seq 1 10); do
    if curl -s http://localhost:$PORT/health > /dev/null 2>&1; then
        echo "Server started on port $PORT (PID: $(cat /tmp/narrator-ai.pid))"
        if [ -n "$OPENROUTER_API_KEY" ]; then
            echo "LLM: configured"
        else
            echo "LLM: NOT configured (set OPENROUTER_API_KEY in .env)"
        fi
        exit 0
    fi
    sleep 1
done

echo "FAILED to start. Check /tmp/narrator-ai.log"
exit 1
