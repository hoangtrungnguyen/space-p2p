#!/bin/bash

# Configuration
API_URL="http://localhost:8080"
ROOM_NAME="demo-room-$(date +%s)"
IDENTITY="user-$(date +%s)"

echo "🚀 Testing Space-P2P API..."
echo "-----------------------------------"

# 1. Create a Room
echo "1. Creating room: $ROOM_NAME"
CREATE_RES=$(curl -s -X POST "$API_URL/rooms" \
  -H "Content-Type: application/json" \
  -d "{\"room_name\": \"$ROOM_NAME\", \"empty_timeout\": 300}")
echo "Response: $CREATE_RES"
echo ""

# 2. List Rooms
echo "2. Listing all rooms"
LIST_RES=$(curl -s -X GET "$API_URL/rooms")
echo "Response: $LIST_RES"
echo ""

# 3. Generate a Token
echo "3. Generating token for $IDENTITY in $ROOM_NAME"
TOKEN_RES=$(curl -s -X POST "$API_URL/token" \
  -H "Content-Type: application/json" \
  -d "{\"room_name\": \"$ROOM_NAME\", \"identity\": \"$IDENTITY\"}")
echo "Response: $TOKEN_RES"
echo ""

# 4. Delete the Room
echo "4. Deleting room: $ROOM_NAME"
DELETE_RES=$(curl -s -X DELETE "$API_URL/rooms/$ROOM_NAME")
echo "Response: $DELETE_RES"
echo ""

echo "-----------------------------------"
echo "✅ Test script completed!"
