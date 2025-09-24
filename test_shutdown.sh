#!/bin/bash

echo "Testing HTTP server graceful shutdown..."

# Build the server
go build -o server .
if [ $? -ne 0 ]; then
    echo "Failed to build server"
    exit 1
fi

# Start server in background
./server &
SERVER_PID=$!

# Wait a moment for server to start
sleep 1

# Check if server is running
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "Server failed to start"
    exit 1
fi

echo "Server started with PID: $SERVER_PID"

# Test that server responds
curl -s http://localhost:8080/health > /dev/null
if [ $? -eq 0 ]; then
    echo "✓ Server is responding to requests"
else
    echo "✗ Server is not responding"
fi

# Send SIGTERM for graceful shutdown
echo "Sending SIGTERM for graceful shutdown..."
kill -TERM $SERVER_PID

# Wait for graceful shutdown (max 5 seconds)
for i in {1..5}; do
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo "✓ Server shut down gracefully in ${i} seconds"
        break
    fi
    sleep 1
done

# Check if server is still running
if kill -0 $SERVER_PID 2>/dev/null; then
    echo "✗ Server did not shut down gracefully, forcing kill"
    kill -9 $SERVER_PID
    exit 1
else
    echo "✓ Graceful shutdown test completed successfully"
fi

# Clean up
rm -f server