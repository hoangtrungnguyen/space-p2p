    #!/bin/bash
set -e

# ==============================================================================
# Space-P2P LOCAL Deployment Test
# ==============================================================================
# Tests the full deployment flow on your local machine (no SSH needed).
# Usage: ./deployment_local.sh
# ==============================================================================

DEPLOY_DIR="$HOME/space-p2p-deploy"
SERVICE_NAME="space-p2p"
BINARY_NAME="space-p2p-server"

echo "============================================"
echo "  Space-P2P Local Deployment Test"
echo "  Deploy Dir: $DEPLOY_DIR"
echo "============================================"
echo ""

# ==============================================================================
# STEP 1: Cross-compile the Go binary for Linux amd64
# ==============================================================================
echo "🔨 Step 1: Building Linux binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$BINARY_NAME" main.go
echo "   ✅ Built: $BINARY_NAME ($(du -h $BINARY_NAME | cut -f1))"
echo ""

# ==============================================================================
# STEP 2: Copy files to deploy directory
# ==============================================================================
echo "📦 Step 2: Copying files to $DEPLOY_DIR..."
mkdir -p "$DEPLOY_DIR"
cp "$BINARY_NAME" .env livekit.yaml "$DEPLOY_DIR/"
chmod +x "$DEPLOY_DIR/$BINARY_NAME"
echo "   ✅ Files copied"
ls -lh "$DEPLOY_DIR/"
echo ""

# ==============================================================================
# STEP 3: Setup LiveKit in Docker
# ==============================================================================
echo "🐳 Step 3: Setting up LiveKit Docker container..."

# Stop existing container if running
docker rm -f livekit-server 2>/dev/null || true

docker run -d \
    --name livekit-server \
    --restart unless-stopped \
    -p 7880:7880 \
    -p 7881:7881 \
    -p 50000-50050:50000-50050/udp \
    -v "$DEPLOY_DIR/livekit.yaml:/etc/livekit.yaml" \
    livekit/livekit-server:latest \
    --config /etc/livekit.yaml

echo "   ✅ LiveKit container started"
echo ""

# ==============================================================================
# STEP 4: Start the Go app (as a background process for testing)
# ==============================================================================
echo "⚙️  Step 4: Starting the Go app..."

# Kill any existing instance
pkill -f "$BINARY_NAME" 2>/dev/null || true
sleep 1

# Run from the deploy directory, loading .env
cd "$DEPLOY_DIR"
./"$BINARY_NAME" &
APP_PID=$!
cd - > /dev/null

echo "   ✅ App started (PID: $APP_PID)"
echo ""

# ==============================================================================
# STEP 5: Verify deployment
# ==============================================================================
echo "🔍 Step 5: Verifying deployment..."
sleep 3

echo ""
echo "   --- LiveKit Container ---"
docker ps --filter name=livekit-server --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "   --- Quick API Test ---"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/rooms 2>/dev/null || echo "failed")
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ API responding (HTTP $HTTP_CODE)"
else
    echo "   ⚠️  API returned HTTP $HTTP_CODE (may still be starting)"
fi

echo ""
echo "   --- Create Room Test ---"
curl -s -X POST http://localhost:8080/rooms \
    -H "Content-Type: application/json" \
    -d '{"room_name": "deploy-test-room"}' | head -c 200
echo ""

echo ""
echo "   --- Generate Token Test ---"
curl -s -X POST http://localhost:8080/token \
    -H "Content-Type: application/json" \
    -d '{"room_name": "deploy-test-room", "identity": "deploy-tester"}' | head -c 200
echo ""

echo ""
echo "   --- Delete Room Test ---"
curl -s -X DELETE http://localhost:8080/rooms/deploy-test-room | head -c 200
echo ""

echo ""
echo "============================================"
echo "  ✅ Local Deployment Test Complete!"
echo "============================================"
echo ""
echo "  App PID: $APP_PID"
echo "  To stop:  kill $APP_PID && docker rm -f livekit-server"
echo "  Cleanup:  rm -rf $DEPLOY_DIR"
echo "============================================"
