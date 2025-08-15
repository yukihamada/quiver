#!/bin/bash
set -e

echo "Using simplified gateways for now..."

GATEWAY_INSTANCES=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(name)")

for INSTANCE in $GATEWAY_INSTANCES; do
    echo "Configuring $INSTANCE..."
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Stop current service
systemctl stop quiver-gateway

# Use the simple gateway that's already built
cat > /etc/systemd/system/quiver-gateway.service << 'EOF'
[Unit]
Description=QUIVer Simple Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
ExecStart=/usr/local/bin/quiver-gateway-simple
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

sleep 1
systemctl status quiver-gateway --no-pager | head -5
"
done

echo ""
echo "Testing gateways..."
sleep 3

for INSTANCE in $GATEWAY_INSTANCES; do
    IP=$(gcloud compute instances describe $INSTANCE --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
    echo ""
    echo "Testing $INSTANCE ($IP):"
    curl -s http://$IP:8080/health | jq . || echo "Failed"
done