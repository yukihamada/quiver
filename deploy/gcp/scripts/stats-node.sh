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

# Install dependencies for realtime stats
cd gateway
go get github.com/gorilla/websocket

# Build realtime stats server
go build -o /usr/local/bin/quiver-stats ./cmd/realtime-stats

# Get bootstrap peer ID
sleep 30
BOOTSTRAP_PEERID=$(curl -s http://${BOOTSTRAP_IP}:8090/health | grep -o '"peer_id":"[^"]*' | cut -d'"' -f4)

# Create systemd service
cat > /etc/systemd/system/quiver-stats.service << EOF
[Unit]
Description=QUIVer Realtime Stats Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
ExecStart=/usr/local/bin/quiver-stats --http :8087 --p2p 4005 --bootstrap /ip4/${BOOTSTRAP_IP}/tcp/4001/p2p/${BOOTSTRAP_PEERID}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Configure nginx as reverse proxy (optional)
apt-get install -y nginx

cat > /etc/nginx/sites-available/quiver-stats << EOF
server {
    listen 80;
    server_name _;

    location /ws {
        proxy_pass http://localhost:8087;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
    }

    location /api/stats {
        proxy_pass http://localhost:8087;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        
        # CORS headers
        add_header Access-Control-Allow-Origin *;
        add_header Access-Control-Allow-Methods "GET, OPTIONS";
        add_header Access-Control-Allow-Headers "Content-Type";
    }
}
EOF

ln -sf /etc/nginx/sites-available/quiver-stats /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default

# Start services
systemctl daemon-reload
systemctl enable quiver-stats
systemctl start quiver-stats
systemctl restart nginx

echo "Stats node setup complete"