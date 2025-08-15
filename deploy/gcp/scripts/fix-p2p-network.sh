#!/bin/bash
set -e

echo "Fixing P2P network connections..."

# Fix bootstrap node to listen on correct port
echo "Configuring bootstrap node..."
gcloud compute ssh root@quiver-bootstrap --zone=asia-northeast1-a --command="
# Stop current service
systemctl stop quiver-bootstrap || true

# Create proper bootstrap configuration
cat > /opt/quiver/bootstrap-config.yaml << 'EOF'
port: 4001
metrics_port: 8090
network:
  listen_addresses:
    - /ip4/0.0.0.0/tcp/4001
    - /ip4/0.0.0.0/udp/4001/quic
  announce_addresses:
    - /ip4/34.85.126.98/tcp/4001
    - /ip4/34.85.126.98/udp/4001/quic
  dht:
    mode: server
  relay:
    enabled: true
    mode: server
EOF

# Create updated service
cat > /etc/systemd/system/quiver-bootstrap.service << 'SVCEOF'
[Unit]
Description=QUIVer Bootstrap Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment=\"PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"
Environment=\"GOPATH=/root/go\"
Environment=\"GOMODCACHE=/root/go/pkg/mod\"
ExecStart=/usr/local/bin/quiver-bootstrap --config /opt/quiver/bootstrap-config.yaml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SVCEOF

# Ensure binary exists
if [ ! -f /usr/local/bin/quiver-bootstrap ]; then
    echo 'Bootstrap binary not found, using provider in bootstrap mode'
    cd /opt/quiver/provider
    go build -o /usr/local/bin/quiver-bootstrap ./cmd/provider
fi

# Start service
systemctl daemon-reload
systemctl enable quiver-bootstrap
systemctl start quiver-bootstrap

# Get peer ID
sleep 5
PEER_ID=\$(journalctl -u quiver-bootstrap -n 100 --no-pager | grep -oE '12D3Koo[a-zA-Z0-9]+' | head -1)
echo \"Bootstrap node Peer ID: \$PEER_ID\"
"

# Get bootstrap peer ID
BOOTSTRAP_IP=$(gcloud compute instances describe quiver-bootstrap --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
echo "Bootstrap node IP: $BOOTSTRAP_IP"

# Update other nodes to connect to bootstrap
echo "Updating provider nodes..."
PROVIDER_IPS=$(gcloud compute instances list --filter="name:quiver-provider-*" --format="value(name)")

for INSTANCE in $PROVIDER_IPS; do
    echo "Updating $INSTANCE..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Update environment to connect to bootstrap
echo 'export DHT_BOOTSTRAP_PEERS=/ip4/$BOOTSTRAP_IP/tcp/4001' >> /etc/environment

# Restart service if exists
systemctl restart quiver-provider || true
" || echo "Failed to update $INSTANCE"
done

# Update gateway nodes
echo "Updating gateway nodes..."
GATEWAY_IPS=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(name)")

for INSTANCE in $GATEWAY_IPS; do
    echo "Updating $INSTANCE..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Update environment to connect to bootstrap
echo 'export DHT_BOOTSTRAP_PEERS=/ip4/$BOOTSTRAP_IP/tcp/4001' >> /etc/environment

# For now gateways are running mock service, skip P2P update
echo 'Gateway $INSTANCE configured for bootstrap: $BOOTSTRAP_IP'
" || echo "Failed to update $INSTANCE"
done

echo "P2P network configuration complete!"
echo ""
echo "Bootstrap node: $BOOTSTRAP_IP:4001"
echo "To verify P2P connections, check logs with:"
echo "  gcloud compute ssh quiver-bootstrap --zone=asia-northeast1-a --command='journalctl -u quiver-bootstrap -f'"