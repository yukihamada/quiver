#!/bin/bash
set -e

# Log all output
exec > >(tee -a /var/log/quiver-bootstrap-setup.log)
exec 2>&1

echo "Starting QUIVer Bootstrap Node setup at $(date)"

# Update system
apt-get update
apt-get install -y git build-essential wget curl

# Install Go 1.23
echo "Installing Go 1.23..."
wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
rm go1.23.0.linux-amd64.tar.gz

# Set up Go environment
export PATH=/usr/local/go/bin:$PATH
export GOPATH=/root/go
export GOMODCACHE=/root/go/pkg/mod
echo 'export PATH=/usr/local/go/bin:$PATH' >> /etc/profile
echo 'export GOPATH=/root/go' >> /etc/profile
echo 'export GOMODCACHE=/root/go/pkg/mod' >> /etc/profile

# Create Go directories
mkdir -p /root/go/{bin,pkg,src}

# Clone QUIVer repository
echo "Cloning QUIVer repository..."
cd /opt
git clone https://github.com/yukihamada/quiver.git
cd quiver

# Build bootstrap node
echo "Building bootstrap node..."
cd provider
go mod download
go build -o /usr/local/bin/quiver-bootstrap ./cmd/bootstrap

# Create systemd service
cat > /etc/systemd/system/quiver-bootstrap.service << EOF
[Unit]
Description=QUIVer Bootstrap Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
Environment="GOPATH=/root/go"
Environment="GOMODCACHE=/root/go/pkg/mod"
ExecStart=/usr/local/bin/quiver-bootstrap --port 4001 --metrics-port 8090
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Start service
echo "Starting QUIVer bootstrap service..."
systemctl daemon-reload
systemctl enable quiver-bootstrap
systemctl start quiver-bootstrap

# Wait for service to start
sleep 10

# Check service status
systemctl status quiver-bootstrap --no-pager || true

# Log the peer ID
echo "Getting peer ID..."
journalctl -u quiver-bootstrap -n 50 --no-pager | grep -E "(PeerID|peer ID|started|listening)" || true

# Open firewall ports
echo "Configuring firewall..."
ufw allow 4001/tcp
ufw allow 4001/udp
ufw allow 8090/tcp
ufw allow 22/tcp
ufw --force enable

echo "Bootstrap node setup complete at $(date)"
echo "Logs available at: /var/log/quiver-bootstrap-setup.log"