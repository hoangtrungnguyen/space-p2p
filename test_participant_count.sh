#!/bin/bash

BASE_URL="http://localhost:8080"
ROOM_NAME="count-test-room-$(date +%s)"
IDENTITY="user-count"

echo "Testing Participant Counts..."

# 1. Create a Room
echo -e "\n1. Creating room: $ROOM_NAME"
curl -s -X POST "$BASE_URL/rooms" \
  -H "Content-Type: application/json" \
  -d "{\"room_name\": \"$ROOM_NAME\"}"

# 2. Get Count (Should be 0)
echo -e "\n2. Getting count for $ROOM_NAME (expect 0)..."
curl -s "$BASE_URL/rooms/$ROOM_NAME/participants/count" | jq .

# 3. Canvas Space Test
CANVAS_ID="space-count-test"
echo -e "\n3. Creating Canvas Space Mapping for $CANVAS_ID..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_ID\", \"identity\": \"$IDENTITY\"}")
echo $RESPONSE


# 4. Get Canvas Count (Should be 0)
echo -e "\n4. Getting count for canvas space $CANVAS_ID (expect 0)..."
COUNT_0=$(curl -s "$BASE_URL/canvas-spaces/$CANVAS_ID/participants/count" | jq -r '.count')
echo "Count: $COUNT_0"

if [ "$COUNT_0" != "0" ]; then
    echo "FAILED: Expected 0, got $COUNT_0"
    exit 1
fi

# 5. Simulate Participant Connection
echo -e "\n5. Simulating a participant connection..."

# Extract token
TOKEN=$(echo $RESPONSE | jq -r '.token')
if [ -z "$TOKEN" ] || [ "$TOKEN" == "null" ]; then
    echo "FAILED: No token found."
    exit 1
fi

# Connect simulator in background
echo "Connecting simulated client..."
# Change localhost:7880 if needed, assuming default dev setup
LIVEKIT_HOST="http://localhost:7880"

# Create a unique log file for this run
LOG_FILE="simulator_$(date +%s).log"

# Use go run to execute the simulator tool
go run tools/simulator.go "$LIVEKIT_HOST" "$TOKEN" > "$LOG_FILE" 2>&1 &
SIM_PID=$!

echo "Waiting 5 seconds for connection..."
sleep 5

# Check if simulator is still running
if ! ps -p $SIM_PID > /dev/null; then
    echo "FAILED: Simulator exited prematurely. Check log:"
    cat "$LOG_FILE"
    rm "$LOG_FILE"
    exit 1
fi

# 6. Verify Count Increase
echo -e "\n6. Getting count for canvas space $CANVAS_ID (expect 1)..."
COUNT_RES=$(curl -s "$BASE_URL/canvas-spaces/$CANVAS_ID/participants/count")
echo "Response: $COUNT_RES"
COUNT_1=$(echo $COUNT_RES | jq -r '.count')

if [ "$COUNT_1" == "1" ]; then
    echo "SUCCESS: Participant count increased to 1."
else
    echo "FAILED: Expected 1, got $COUNT_1"
    echo "Simulator Log:"
    cat "$LOG_FILE"
fi

# 7. Cleanup
echo -e "\n7. Cleanup..."
kill $SIM_PID 2>/dev/null
rm "$LOG_FILE" 2>/dev/null

echo "Done."
