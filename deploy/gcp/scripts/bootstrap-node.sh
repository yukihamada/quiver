#!/bin/bash
set -e

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

# Build bootstrap node
cd provider
go build -o /usr/local/bin/quiver-bootstrap ./cmd/bootstrap

# Create systemd service
cat > /etc/systemd/system/quiver-bootstrap.service << EOF
[Unit]
Description=QUIVer Bootstrap Node
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
ExecStart=/usr/local/bin/quiver-bootstrap --port 4001 --metrics-port 8090
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable quiver-bootstrap
systemctl start quiver-bootstrap

# Log the peer ID
sleep 5
journalctl -u quiver-bootstrap -n 50 | grep "PeerID" || true

echo "Bootstrap node setup complete"