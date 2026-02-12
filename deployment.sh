#!/bin/bash
set -e

# ==============================================================================
# Space-P2P VPS Deployment Script
# ==============================================================================
# This script handles the full deployment of your Go server + LiveKit (Docker)
# to a Linux VPS.
#
# Usage:
#   ./deployment.sh <vps-user> <vps-ip> [deploy-dir]
#
# Example:
#   ./deployment.sh root 192.168.1.100
#   ./deployment.sh ubuntu 10.0.0.5 /opt/space-p2p
# ==============================================================================

# --- Configuration ---
VPS_USER="${1:?Usage: ./deployment.sh <vps-user> <vps-ip> [deploy-dir]}"
VPS_IP="${2:?Usage: ./deployment.sh <vps-user> <vps-ip> [deploy-dir]}"
DEPLOY_DIR="${3:-/home/$VPS_USER/space-p2p}"
SERVICE_NAME="space-p2p"
BINARY_NAME="space-p2p-server"

echo "============================================"
echo "  Space-P2P Deployment"
echo "============================================"
echo "  VPS:        $VPS_USER@$VPS_IP"
echo "  Deploy Dir: $DEPLOY_DIR"
echo "============================================"
echo ""

# ==============================================================================
# STEP 1: Cross-compile the Go binary for Linux amd64
# ==============================================================================
echo "🔨 Step 1: Building Linux binary..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o "$BINARY_NAME" .
echo "   ✅ Built: $BINARY_NAME ($(du -h $BINARY_NAME | cut -f1))"
echo ""

# ==============================================================================
# STEP 2: Transfer files to VPS
# ==============================================================================
echo "📦 Step 2: Uploading files to VPS..."

ssh "$VPS_USER@$VPS_IP" "mkdir -p $DEPLOY_DIR && sudo systemctl stop $SERVICE_NAME || true"

scp -r "$BINARY_NAME" \
    .env \
    livekit.yaml \
    public \
    "$VPS_USER@$VPS_IP:$DEPLOY_DIR/"

echo "   ✅ Files uploaded to $DEPLOY_DIR"
echo ""

# ==============================================================================
# STEP 3: Setup LiveKit Server (Binary) on VPS
# ==============================================================================
echo "🎥 Step 3: Setting up LiveKit Server..."

ssh "$VPS_USER@$VPS_IP" bash -s "$DEPLOY_DIR" << 'REMOTE_LIVEKIT'
DEPLOY_DIR="$1"
LK_VERSION="v1.8.0" # You can update this version
LK_BINARY_URL="https://github.com/livekit/livekit/releases/download/${LK_VERSION}/livekit_${LK_VERSION#v}_linux_amd64.tar.gz"

cd "$DEPLOY_DIR"

# Check if livekit-server binary exists, if not download it
if [ ! -f "livekit-server" ]; then
    echo "   Downloading LiveKit Server ${LK_VERSION}..."
    if command -v curl &> /dev/null; then
        curl -L -o livekit.tar.gz "$LK_BINARY_URL"
    else
        wget -O livekit.tar.gz "$LK_BINARY_URL"
    fi
    
    echo "   Extracting LiveKit Server..."
    tar -xzf livekit.tar.gz
    rm livekit.tar.gz
    chmod +x livekit-server
fi

# Create systemd service for LiveKit
echo "   Creating LiveKit systemd service..."
sudo tee "/etc/systemd/system/livekit-server.service" > /dev/null << EOF
[Unit]
Description=LiveKit Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/livekit-server --config $DEPLOY_DIR/livekit.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Reload and start LiveKit
sudo systemctl daemon-reload
sudo systemctl enable livekit-server
sudo systemctl restart livekit-server

echo "   ✅ LiveKit Server started"
REMOTE_LIVEKIT

echo ""

# ==============================================================================
# STEP 4: Create systemd service for the Go app
# ==============================================================================
echo "⚙️  Step 4: Creating systemd service..."

ssh "$VPS_USER@$VPS_IP" bash -s "$DEPLOY_DIR" "$SERVICE_NAME" "$BINARY_NAME" "$VPS_USER" << 'REMOTE_SERVICE'
DEPLOY_DIR="$1"
SERVICE_NAME="$2"
BINARY_NAME="$3"
RUN_USER="$4"

# Make binary executable
chmod +x "$DEPLOY_DIR/$BINARY_NAME"

# Create systemd service file
sudo tee "/etc/systemd/system/$SERVICE_NAME.service" > /dev/null << EOF
[Unit]
Description=Space P2P Session Manager
After=network.target livekit-server.service
Wants=livekit-server.service

[Service]
Type=simple
User=$RUN_USER
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DEPLOY_DIR/$BINARY_NAME
Restart=always
RestartSec=5
EnvironmentFile=$DEPLOY_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

# Reload and start
sudo systemctl daemon-reload
sudo systemctl enable "$SERVICE_NAME"
sudo systemctl restart "$SERVICE_NAME"

echo "   ✅ Service '$SERVICE_NAME' created and started"
REMOTE_SERVICE

echo ""

# ==============================================================================
# STEP 5: Configure firewall
# ==============================================================================
echo "🔥 Step 5: Configuring firewall..."

ssh "$VPS_USER@$VPS_IP" << 'REMOTE_FIREWALL'
if command -v ufw &> /dev/null; then
    sudo ufw allow 8080/tcp   comment "Space-P2P API"
    sudo ufw allow 7880/tcp   comment "LiveKit HTTP API"
    sudo ufw allow 7881/tcp   comment "LiveKit WebRTC TCP"
    sudo ufw allow 50000:60000/udp comment "LiveKit WebRTC UDP"
    echo "   ✅ UFW rules added"
elif command -v firewall-cmd &> /dev/null; then
    sudo firewall-cmd --permanent --add-port=8080/tcp
    sudo firewall-cmd --permanent --add-port=7880/tcp
    sudo firewall-cmd --permanent --add-port=7881/tcp
    sudo firewall-cmd --permanent --add-port=50000-60000/udp
    sudo firewall-cmd --reload
    echo "   ✅ firewalld rules added"
else
    echo "   ⚠️  No firewall manager (ufw/firewalld) found. Make sure ports 8080, 7880, 7881, 50000-60000/udp are open."
fi
REMOTE_FIREWALL

echo ""

# ==============================================================================
# STEP 6: Verify deployment
# ==============================================================================
echo "🔍 Step 6: Verifying deployment..."

ssh "$VPS_USER@$VPS_IP" bash -s "$SERVICE_NAME" << 'REMOTE_VERIFY'
SERVICE_NAME="$1"

echo "   --- Services Status ---"
sudo systemctl status "$SERVICE_NAME" --no-pager -l | head -10
echo ""
sudo systemctl status livekit-server --no-pager -l | head -10

echo ""
echo "   --- Quick API Test ---"
sleep 2
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/rooms 2>/dev/null || echo "failed")
if [ "$HTTP_CODE" = "200" ]; then
    echo "   ✅ API responding (HTTP $HTTP_CODE)"
else
    echo "   ⚠️  API returned HTTP $HTTP_CODE (may still be starting)"
fi
REMOTE_VERIFY

echo ""
echo "============================================"
echo "  ✅ Deployment Complete!"
echo "============================================"
echo ""
echo "  Your app:     http://$VPS_IP:8080"
echo "  LiveKit:      http://$VPS_IP:7880"
echo ""
echo "  Useful commands (run on VPS):"
echo "    sudo systemctl status $SERVICE_NAME"
echo "    sudo systemctl status livekit-server"
echo "    sudo journalctl -u $SERVICE_NAME -f"
echo "    sudo journalctl -u livekit-server -f"
echo ""
echo "  ⚠️  IMPORTANT: Update your .env on the VPS!"
echo "  Set LIVEKIT_HOST=http://$VPS_IP:7880"
echo "  (or your domain if you have one)"
echo "============================================"
