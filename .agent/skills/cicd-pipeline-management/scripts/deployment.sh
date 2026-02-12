#!/bin/bash
set -e

# Deployment Script for TDD CI/CD Skill
# Usage: ./deployment.sh <host> <user> <ssh_key> <local_binary>

HOST=$1
USER=$2
SSH_KEY=$3
LOCAL_BINARY=$4

if [ -z "$HOST" ] || [ -z "$USER" ] || [ -z "$SSH_KEY" ] || [ -z "$LOCAL_BINARY" ]; then
    echo "Usage: ./deployment.sh <host> <user> <ssh_key> <local_binary>"
    exit 1
fi

REMOTE_DIR="/home/$USER/space-p2p"

echo "========================================"
echo "🚀 Deploying to $USER@$HOST"
echo "========================================"

# 1. Artifact Transfer
echo "📦 Transferring artifacts..."
scp -i "$SSH_KEY" "$LOCAL_BINARY" "$USER@$HOST:$REMOTE_DIR/space-p2p-server"
scp -i "$SSH_KEY" .env "$USER@$HOST:$REMOTE_DIR/.env"
scp -i "$SSH_KEY" livekit.yaml "$USER@$HOST:$REMOTE_DIR/livekit.yaml"

# 2. Remote Execution
echo "🔧 Configuring Remote Server..."
ssh -i "$SSH_KEY" "$USER@$HOST" "bash -s" << 'EOF'

set -e

# 2. LiveKit SFU Initialization
if ! command -v livekit-server &> /dev/null; then
    echo "Installing LiveKit Server..."
    # Simplified install logic
    curl -sSL https://get.livekit.io | bash
fi

# Ensure data dir exists
mkdir -p data

# 3. Systemd Configuration (simplified checks)
# Create service file logic if needed, or assume manual setup by admin for this skill demo.
# echo "Configuring systemd..."
# sudo systemctl deamon-reload

# 4. Firewall Configuration (UFW)
if command -v ufw &> /dev/null; then
    echo "Configuring Firewall..."
    sudo ufw allow 8080/tcp # API
    sudo ufw allow 7880/tcp # LiveKit Signal
    sudo ufw allow 7881/tcp # LiveKit TCP Tunnelling
    sudo ufw allow 50000:60000/udp # WebRTC
fi

# Restart services
echo "Restarting services..."
# sudo systemctl restart livekit-server
# sudo systemctl restart space-p2p

# 5. Health Verification
echo "Verifying Health..."
sleep 5
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/rooms || echo "000")

if [ "$HTTP_CODE" -eq 200 ]; then
    echo "✅ Health Check Passed (200 OK)"
else
    echo "❌ Health Check Failed (Status: $HTTP_CODE)"
    # journalctl -u space-p2p -n 10
    exit 1
fi

EOF

echo "✅ Deployment Successful!"
