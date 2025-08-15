#!/bin/bash
set -e

echo "Switching to real P2P gateways..."

# Get bootstrap info
BOOTSTRAP_IP=$(gcloud compute instances describe quiver-bootstrap --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
BOOTSTRAP_PEER_ID=$(gcloud compute ssh root@quiver-bootstrap --zone=asia-northeast1-a --command="journalctl -u quiver-bootstrap -n 100 --no-pager 2>/dev/null | grep -oE '12D3Koo[a-zA-Z0-9]+' | head -1" 2>/dev/null || echo "12D3KooWBBV3Sp2nCk47ygTCpCYfudkURmJUHK1Zmv33o4hCa991")

echo "Bootstrap: /ip4/$BOOTSTRAP_IP/tcp/4001/p2p/$BOOTSTRAP_PEER_ID"

GATEWAY_INSTANCES=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(name)")

for INSTANCE in $GATEWAY_INSTANCES; do
    echo "Switching $INSTANCE to real P2P gateway..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Stop mock gateway
systemctl stop quiver-gateway

# Create real gateway service
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

# Reload and start real gateway
systemctl daemon-reload
systemctl enable quiver-gateway
systemctl start quiver-gateway

sleep 2
systemctl status quiver-gateway --no-pager | head -10
"
done

echo ""
echo "Testing real P2P gateways..."
sleep 5

for INSTANCE in $GATEWAY_INSTANCES; do
    IP=$(gcloud compute instances describe $INSTANCE --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
    echo ""
    echo "Testing $INSTANCE ($IP):"
    curl -s http://$IP:8080/health | jq . || echo "Failed"
done

echo ""
echo "To test inference:"
echo "curl -X POST http://<gateway-ip>:8080/generate \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"prompt\":\"Hello world\",\"model\":\"llama3.2:3b\",\"token\":\"test\"}'"