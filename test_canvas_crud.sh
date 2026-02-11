#!/bin/bash

BASE_URL="http://localhost:8080"
CANVAS_SPACE_ID="test-space-1"
IDENTITY="user1"

echo "Testing Canvas Space CRUD..."

# 1. Create/Connect (First time) - Should create new room
echo -e "\n1. connecting to $CANVAS_SPACE_ID (First time)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\", \"identity\": \"$IDENTITY\"}")

echo "Response: $RESPONSE"
ROOM_NAME=$(echo $RESPONSE | jq -r '.room_name')
TOKEN=$(echo $RESPONSE | jq -r '.token')

if [ "$ROOM_NAME" == "null" ] || [ -z "$ROOM_NAME" ]; then
  echo "FAILED: Room name not returned."
  exit 1
fi
echo "Created Room: $ROOM_NAME"

# 2. Read Canvas Space
echo -e "\n2. Getting info for $CANVAS_SPACE_ID..."
RESPONSE=$(curl -s -X GET "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
echo "Response: $RESPONSE"
FETCHED_ROOM=$(echo $RESPONSE | jq -r '.room_name')

if [ "$FETCHED_ROOM" != "$ROOM_NAME" ]; then
  echo "FAILED: Fetched room ($FETCHED_ROOM) does not match created room ($ROOM_NAME)."
  exit 1
fi
echo "Verified room mapping."

# 3. Connect Again - Should return SAME room
echo -e "\n3. Connecting to $CANVAS_SPACE_ID (Second time)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\", \"identity\": \"user2\"}")
echo "Response: $RESPONSE"
SECOND_ROOM=$(echo $RESPONSE | jq -r '.room_name')

if [ "$SECOND_ROOM" != "$ROOM_NAME" ]; then
  echo "FAILED: Second connection returned different room ($SECOND_ROOM) vs ($ROOM_NAME)."
  exit 1
fi
echo "Verified consistency."

# 4. List All
echo -e "\n4. Listing all canvas spaces..."
curl -s -X GET "$BASE_URL/canvas-spaces" | jq .

# 5. Delete
echo -e "\n5. Deleting $CANVAS_SPACE_ID..."
curl -s -X DELETE "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID"

# 6. Verify Deletion
echo -e "\n6. Verifying deletion..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
if [ "$HTTP_CODE" != "404" ]; then
    echo "FAILED: Expected 404 after deletion, got $HTTP_CODE"
    exit 1
fi
echo "Verified deletion (404)."

# 7. Connect After Delete - Should create NEW room
echo -e "\n7. Connecting to $CANVAS_SPACE_ID after delete..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces/token" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\", \"identity\": \"user1\"}")
echo "Response: $RESPONSE"
NEW_ROOM=$(echo $RESPONSE | jq -r '.room_name')

if [ "$NEW_ROOM" == "$ROOM_NAME" ]; then
  echo "WARNING: New room name is same as old one. Is random generation working? Or just unlucky?"
else
  echo "Created NEW Room: $NEW_ROOM"
fi

echo -e "\nSUCCESS: All tests passed."
