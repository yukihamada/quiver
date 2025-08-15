#!/bin/bash
set -e

echo "Deploying real P2P QUIVer network..."

# Step 1: Deploy Bootstrap Node with proper P2P setup
echo "Step 1: Configuring Bootstrap Node..."
gcloud compute ssh root@quiver-bootstrap --zone=asia-northeast1-a --command="
# Install Go if not present
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    rm go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=/usr/local/go/bin:\$PATH' >> /etc/profile
    export PATH=/usr/local/go/bin:\$PATH
fi

# Clone and build QUIVer
cd /opt
rm -rf quiver
git clone https://github.com/yukihamada/quiver.git
cd quiver/provider

# Build bootstrap binary
go mod download
go build -o /usr/local/bin/quiver-bootstrap ./cmd/provider

# Create bootstrap service with P2P configuration
cat > /etc/systemd/system/quiver-bootstrap.service << 'EOF'
[Unit]
Description=QUIVer Bootstrap Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment=\"PROVIDER_LISTEN_ADDR=/ip4/0.0.0.0/tcp/4001\"
Environment=\"PROVIDER_METRICS_PORT=8090\"
Environment=\"PROVIDER_MODE=bootstrap\"
ExecStart=/usr/local/bin/quiver-bootstrap
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable quiver-bootstrap
systemctl restart quiver-bootstrap

sleep 5
PEER_ID=\$(journalctl -u quiver-bootstrap -n 100 --no-pager | grep -oE '12D3Koo[a-zA-Z0-9]+' | head -1)
echo \"Bootstrap Peer ID: \$PEER_ID\"
"

BOOTSTRAP_IP=$(gcloud compute instances describe quiver-bootstrap --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
echo "Bootstrap IP: $BOOTSTRAP_IP"

# Wait for bootstrap to get peer ID
sleep 10
BOOTSTRAP_PEER_ID=$(gcloud compute ssh root@quiver-bootstrap --zone=asia-northeast1-a --command="journalctl -u quiver-bootstrap -n 100 --no-pager | grep -oE '12D3Koo[a-zA-Z0-9]+' | head -1" 2>/dev/null || echo "")

if [ -z "$BOOTSTRAP_PEER_ID" ]; then
    echo "Warning: Could not get bootstrap peer ID"
    BOOTSTRAP_PEER_ID="PEER_ID_PLACEHOLDER"
fi

echo "Bootstrap Peer: /ip4/$BOOTSTRAP_IP/tcp/4001/p2p/$BOOTSTRAP_PEER_ID"

# Step 2: Deploy Provider Nodes with Ollama
echo "Step 2: Deploying Provider Nodes..."
PROVIDER_INSTANCES=$(gcloud compute instances list --filter="name:quiver-provider-*" --format="value(name)")

for INSTANCE in $PROVIDER_INSTANCES; do
    echo "Configuring $INSTANCE..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Install Go if not present
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    rm go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=/usr/local/go/bin:\$PATH' >> /etc/profile
    export PATH=/usr/local/go/bin:\$PATH
fi

# Install Ollama
if ! command -v ollama &> /dev/null; then
    curl -fsSL https://ollama.ai/install.sh | sh
fi

# Start Ollama service
systemctl enable ollama
systemctl start ollama

# Pull model
ollama pull llama3.2:3b || true

# Clone and build QUIVer provider
cd /opt
rm -rf quiver
git clone https://github.com/yukihamada/quiver.git
cd quiver/provider

# Build provider binary
go mod download
go build -o /usr/local/bin/quiver-provider ./cmd/provider

# Create provider service
cat > /etc/systemd/system/quiver-provider.service << 'EOF'
[Unit]
Description=QUIVer Provider Node
After=network-online.target ollama.service
Wants=network-online.target
Requires=ollama.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment=\"PROVIDER_LISTEN_ADDR=/ip4/0.0.0.0/tcp/4002\"
Environment=\"PROVIDER_DHT_BOOTSTRAP_PEERS=/ip4/$BOOTSTRAP_IP/tcp/4001/p2p/$BOOTSTRAP_PEER_ID\"
Environment=\"PROVIDER_OLLAMA_URL=http://localhost:11434\"
Environment=\"PROVIDER_METRICS_PORT=8091\"
ExecStart=/usr/local/bin/quiver-provider
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable quiver-provider
systemctl restart quiver-provider

echo \"Provider $INSTANCE configured\"
" || echo "Failed to configure $INSTANCE"
done

# Step 3: Deploy Real Gateway with P2P Client
echo "Step 3: Deploying Real P2P Gateways..."
GATEWAY_INSTANCES=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(name)")

for INSTANCE in $GATEWAY_INSTANCES; do
    echo "Configuring $INSTANCE..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Stop mock gateway
systemctl stop quiver-gateway || true

# Install Go if not present
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
    tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
    rm go1.23.0.linux-amd64.tar.gz
    echo 'export PATH=/usr/local/go/bin:\$PATH' >> /etc/profile
    export PATH=/usr/local/go/bin:\$PATH
fi

# Clone and build QUIVer gateway
cd /opt
rm -rf quiver
git clone https://github.com/yukihamada/quiver.git
cd quiver/gateway

# Build gateway binary
go mod download
go build -o /usr/local/bin/quiver-gateway ./cmd/gateway

# Create gateway service
cat > /etc/systemd/system/quiver-gateway.service << 'EOF'
[Unit]
Description=QUIVer P2P Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment=\"GATEWAY_PORT=8080\"
Environment=\"GATEWAY_P2P_LISTEN=/ip4/0.0.0.0/tcp/4003\"
Environment=\"GATEWAY_DHT_BOOTSTRAP_PEERS=/ip4/$BOOTSTRAP_IP/tcp/4001/p2p/$BOOTSTRAP_PEER_ID\"
Environment=\"GATEWAY_RATE_LIMIT_PER_TOKEN=10\"
Environment=\"GATEWAY_CANARY_RATE=0.1\"
ExecStart=/usr/local/bin/quiver-gateway
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable quiver-gateway
systemctl restart quiver-gateway

echo \"Gateway $INSTANCE configured\"
" || echo "Failed to configure $INSTANCE"
done

echo ""
echo "Deployment complete! Checking status..."
sleep 10

# Check status
echo ""
echo "=== Bootstrap Node Status ==="
gcloud compute ssh root@quiver-bootstrap --zone=asia-northeast1-a --command="systemctl status quiver-bootstrap --no-pager | head -20" 2>/dev/null || echo "Failed to check bootstrap status"

echo ""
echo "=== Provider Status ==="
for INSTANCE in $PROVIDER_INSTANCES; do
    echo "--- $INSTANCE ---"
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="systemctl status quiver-provider --no-pager | head -10; ollama list" 2>/dev/null || echo "Failed to check $INSTANCE"
done

echo ""
echo "=== Gateway Status ==="
for INSTANCE in $GATEWAY_INSTANCES; do
    IP=$(gcloud compute instances describe $INSTANCE --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
    echo "--- $INSTANCE ($IP) ---"
    curl -s http://$IP:8080/health | jq . || echo "Health check failed"
done

echo ""
echo "P2P Network Information:"
echo "Bootstrap: /ip4/$BOOTSTRAP_IP/tcp/4001/p2p/$BOOTSTRAP_PEER_ID"
echo ""
echo "Test with:"
echo "curl -X POST http://<gateway-ip>:8080/generate -H 'Content-Type: application/json' -d '{\"prompt\":\"Hello\",\"model\":\"llama3.2:3b\",\"token\":\"test\"}'"