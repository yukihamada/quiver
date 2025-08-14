#!/bin/bash
set -e

# Get bootstrap address from metadata
BOOTSTRAP_IP=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/bootstrap_addr" -H "Metadata-Flavor: Google")

# Update system
apt-get update
apt-get install -y git build-essential wget curl

# Install Go
wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile

# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Start Ollama service
systemctl start ollama
systemctl enable ollama

# Pull llama3.2 model
ollama pull llama3.2:3b

# Clone QUIVer repository
cd /opt
git clone https://github.com/yukihamada/quiver.git
cd quiver

# Build provider
cd provider
go build -o /usr/local/bin/quiver-provider ./cmd/provider

# Get bootstrap peer ID (wait for it to be ready)
sleep 30
BOOTSTRAP_PEERID=$(curl -s http://${BOOTSTRAP_IP}:8090/health | grep -o '"peer_id":"[^"]*' | cut -d'"' -f4)

# Create systemd service
cat > /etc/systemd/system/quiver-provider.service << EOF
[Unit]
Description=QUIVer Provider Node
After=network.target ollama.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
ExecStart=/usr/local/bin/quiver-provider --bootstrap /ip4/${BOOTSTRAP_IP}/tcp/4001/p2p/${BOOTSTRAP_PEERID}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
Environment="OLLAMA_URL=http://localhost:11434"

[Install]
WantedBy=multi-user.target
EOF

# Start service
systemctl daemon-reload
systemctl enable quiver-provider
systemctl start quiver-provider

echo "Provider node setup complete"