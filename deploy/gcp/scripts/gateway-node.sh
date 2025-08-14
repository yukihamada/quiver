#!/bin/bash
set -e

# Get bootstrap address from metadata
BOOTSTRAP_IP=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/bootstrap_addr" -H "Metadata-Flavor: Google")

# Update system
apt-get update
apt-get install -y git build-essential wget

# Install Go
wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Clone QUIVer repository
cd /opt
git clone https://github.com/yukihamada/quiver.git
cd quiver

# Build gateway
cd gateway
go build -o /usr/local/bin/quiver-gateway ./cmd/gateway

# Get bootstrap peer ID
sleep 30
BOOTSTRAP_PEERID=$(curl -s http://${BOOTSTRAP_IP}:8090/health | grep -o '"peer_id":"[^"]*' | cut -d'"' -f4)

# Create systemd service
cat > /etc/systemd/system/quiver-gateway.service << EOF
[Unit]
Description=QUIVer Gateway Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
ExecStart=/usr/local/bin/quiver-gateway --bootstrap /ip4/${BOOTSTRAP_IP}/tcp/4001/p2p/${BOOTSTRAP_PEERID} --port 8081
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable quiver-gateway
systemctl start quiver-gateway

echo "Gateway node setup complete"