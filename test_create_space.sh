#!/bin/bash

# Configuration
BASE_URL="http://localhost:8080"
CANVAS_SPACE_ID="test-space-$(date +%s)"

echo "------------------------------------------------"
echo "Testing Create Canvas Space: $CANVAS_SPACE_ID"
echo "URL: $BASE_URL/canvas-spaces"
echo "------------------------------------------------"

# 1. Create Canvas Space
echo "Step 1: Creating canvas space..."
RESPONSE=$(curl -s -X POST "$BASE_URL/canvas-spaces" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\"}")

echo "Response Content: $RESPONSE"

# Check if mapping was created using jq if available, otherwise raw check
ROOM_NAME=$(echo $RESPONSE | jq -r '.room_name' 2>/dev/null)

if [ "$ROOM_NAME" != "" ] && [ "$ROOM_NAME" != "null" ]; then
    echo "SUCCESS: Created mapping to Room: $ROOM_NAME"
else
    echo "ERROR: Failed to create canvas space. Full response above."
    exit 1
fi

# 2. Try Duplicate (Should fail with 409)
echo -e "\nStep 2: Testing duplicate creation (should fail)..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/canvas-spaces" \
  -H "Content-Type: application/json" \
  -d "{\"canvas_space_id\": \"$CANVAS_SPACE_ID\"}")

if [ "$HTTP_CODE" == "409" ]; then
    echo "SUCCESS: Received expected 409 Conflict code."
else
    echo "ERROR: Expected 409 Conflict, but got $HTTP_CODE"
fi

# 3. Verify via GET
echo -e "\nStep 3: Verifying space exists via GET..."
GET_RESPONSE=$(curl -s -X GET "$BASE_URL/canvas-spaces/$CANVAS_SPACE_ID")
echo "GET Response: $GET_RESPONSE"

echo -e "\n------------------------------------------------"
echo "Test complete for ID: $CANVAS_SPACE_ID"
echo "------------------------------------------------"
