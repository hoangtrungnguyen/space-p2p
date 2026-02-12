#!/bin/bash

# Configuration for Production
BASE_URL="http://34.143.219.209:8080"
CANVAS_SPACE_ID="prod-crud-test-$(date +%s)"
IDENTITY="prod-admin"

echo "🚀 Running Production CRUD Tests for Canvas Spaces..."
echo "Target URL: $BASE_URL"
echo "Space ID:   $CANVAS_SPACE_ID"
echo "--------------------------------------------------------"

# 1. Create/Connect (First time) - Should create new room mapping
echo -e "\n1. Connecting (First time)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\", \"identity\": \"$IDENTITY\"}")

echo "Response: $RESPONSE"
ROOM_NAME=$(echo $RESPONSE | grep -o '"room_name":"[^"]*' | cut -d'"' -f4)

if [ -z "$ROOM_NAME" ]; then
  echo "❌ FAILED: Room name not returned."
  exit 1
fi
echo "✅ Created Room Mapping: $ROOM_NAME"

# 2. Read Canvas Space
echo -e "\n2. Verifying mapping exists..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
echo "Response: $GET_RESPONSE"
FETCHED_ROOM=$(echo $GET_RESPONSE | grep -o '"room_name":"[^"]*' | cut -d'"' -f4)

if [ "$FETCHED_ROOM" != "$ROOM_NAME" ]; then
  echo "❌ FAILED: Fetched room ($FETCHED_ROOM) does not match created room ($ROOM_NAME)."
  exit 1
fi
echo "✅ Verified room mapping."

# 3. List All Canvas Spaces
echo -e "\n3. Listing all canvas spaces..."
LIST_RESPONSE=$(curl -s -X GET "$BASE_URL/canvas-spaces")
if [[ $LIST_RESPONSE == *"$CANVAS_SPACE_ID"* ]]; then
    echo "✅ Success: Found $CANVAS_SPACE_ID in list"
else
    echo "❌ FAILED: $CANVAS_SPACE_ID not found in list"
    echo "Response: $LIST_RESPONSE"
    exit 1
fi

# 4. Check Participant Count
echo -e "\n4. Checking participant count..."
COUNT_RESPONSE=$(curl -s -X GET "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID/participants/count")
echo "Response: $COUNT_RESPONSE"
if [[ $COUNT_RESPONSE == *"count"* ]]; then
    echo "✅ Success: Received count info"
else
    echo "❌ FAILED: Invalid count response"
    exit 1
fi

# 5. Delete Mapping
echo -e "\n5. Deleting $CANVAS_SPACE_ID..."
DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
echo "Response: $DELETE_RESPONSE"

# 6. Verify Deletion (404)
echo -e "\n6. Verifying deletion (should be 404)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
if [ "$HTTP_CODE" == "404" ]; then
    echo "✅ Success: Mapping is gone (404)"
else
    echo "❌ FAILED: Expected 404, got $HTTP_CODE"
    exit 1
fi

echo -e "\n--------------------------------------------------------"
echo "🏁 Production CRUD tests passed successfully!"
