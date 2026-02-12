#!/bin/bash

# Configuration for Production
BASE_URL="http://34.143.219.209:8080"
CANVAS_SPACE_ID="prod-test-space-$(date +%s)"
IDENTITY="prod-tester-$(date +%s)"

echo "🚀 Running Production Connectivity Tests..."
echo "Target URL: $BASE_URL"
echo "------------------------------------------------"

# 1. Health Check (Root Redirect)
echo -e "\n1. Testing Root Health/Redirect..."
STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/")
if [ "$STATUS" == "301" ] || [ "$STATUS" == "200" ]; then
    echo "✅ Success (HTTP $STATUS)"
else
    echo "❌ Failed (HTTP $STATUS)"
fi

# 2. List Canvas Spaces
echo -e "\n2. Listing Canvas Spaces..."
LIST_RES=$(curl -s GET "$BASE_URL/canvas-spaces")
if [[ $LIST_RES == *"["* ]]; then
    echo "✅ Success: API returned list"
else
    echo "❌ Failed: Invalid response"
    echo "Response: $LIST_RES"
fi

# 3. Create/Connect Canvas Space
echo -e "\n3. Creating Canvas Space Connection..."
CONN_RES=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\", \"identity\": \"$IDENTITY\"}")

ROOM_NAME=$(echo $CONN_RES | grep -o '"room_name":"[^"]*' | cut -d'"' -f4)
TOKEN=$(echo $CONN_RES | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ ! -z "$ROOM_NAME" ] && [ ! -z "$TOKEN" ]; then
    echo "✅ Success: Created mapping to $ROOM_NAME"
    echo "✅ Success: Received JWT Token"
else
    echo "❌ Failed: Could not get room_name or token"
    echo "Response: $CONN_RES"
fi

# 4. Check Participant Count
echo -e "\n4. Checking Participant Count..."
COUNT_RES=$(curl -s GET "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID/participants/count")
if [[ $COUNT_RES == *"count"* ]]; then
    echo "✅ Success: API returned participant count"
    echo "Response: $COUNT_RES"
else
    echo "❌ Failed"
    echo "Response: $COUNT_RES"
fi

# 5. Cleanup (Optional)
echo -e "\n5. Cleaning up test mapping..."
DELETE_RES=$(curl -s -X DELETE "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
if [[ $DELETE_RES == *"deleted"* ]]; then
    echo "✅ Success: Cleaned up $CANVAS_SPACE_ID"
else
    echo "❌ Cleanup Failed"
fi

echo -e "\n------------------------------------------------"
echo "🏁 Production tests completed!"
