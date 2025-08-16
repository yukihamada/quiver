#!/bin/bash

set -e

echo "🔐 QUIVer Secure Deployment Script"
echo "================================="

# Check for required environment
if [ ! -f .env ]; then
    echo "❌ .env file not found!"
    echo "Please copy .env.example to .env and configure it."
    exit 1
fi

# Load environment
source .env

# Validate critical settings
if [ "$QUIVER_JWT_SECRET" = "your-secure-jwt-secret-here" ]; then
    echo "❌ Please set a secure JWT secret in .env!"
    exit 1
fi

# Generate secure secrets if not set
if [ -z "$QUIVER_JWT_SECRET" ]; then
    echo "🔑 Generating secure JWT secret..."
    JWT_SECRET=$(openssl rand -base64 32)
    echo "QUIVER_JWT_SECRET=$JWT_SECRET" >> .env
fi

# Build with security flags
echo "🔨 Building with security hardening..."
export CGO_ENABLED=1
export GOFLAGS="-buildmode=pie"

# Build all components
make build-all

# Generate checksums
echo "📝 Generating checksums..."
./scripts/generate-checksums.sh

# Create secure directories
echo "📁 Creating secure directories..."
sudo mkdir -p /etc/quiver/{tls,keys}
sudo chmod 700 /etc/quiver/keys

# Generate TLS certificates if needed
if [ "$ENABLE_TLS" = "true" ] && [ ! -f "$TLS_CERT_PATH" ]; then
    echo "🔒 Generating TLS certificates..."
    sudo openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
        -keyout "$TLS_KEY_PATH" \
        -out "$TLS_CERT_PATH" \
        -subj "/C=US/ST=State/L=City/O=QUIVer/CN=quiver.network"
    sudo chmod 600 "$TLS_KEY_PATH"
    sudo chmod 644 "$TLS_CERT_PATH"
fi

# Create systemd services
echo "🚀 Creating systemd services..."

# Gateway service
sudo tee /etc/systemd/system/quiver-gateway.service > /dev/null << EOF
[Unit]
Description=QUIVer Gateway Service
After=network.target

[Service]
Type=simple
User=quiver
Group=quiver
EnvironmentFile=/etc/quiver/gateway.env
ExecStart=/usr/local/bin/quiver-gateway
Restart=always
RestartSec=10
StandardOutput=append:/var/log/quiver/gateway.log
StandardError=append:/var/log/quiver/gateway.error.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/quiver
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
RestrictNamespaces=true
LockPersonality=true
MemoryDenyWriteExecute=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true

[Install]
WantedBy=multi-user.target
EOF

# Provider service
sudo tee /etc/systemd/system/quiver-provider.service > /dev/null << EOF
[Unit]
Description=QUIVer Provider Service
After=network.target ollama.service

[Service]
Type=simple
User=quiver
Group=quiver
EnvironmentFile=/etc/quiver/provider.env
ExecStart=/usr/local/bin/quiver-provider
Restart=always
RestartSec=10
StandardOutput=append:/var/log/quiver/provider.log
StandardError=append:/var/log/quiver/provider.error.log

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/quiver /var/lib/quiver
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectControlGroups=true
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
RestrictNamespaces=true
LockPersonality=true
MemoryDenyWriteExecute=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true

[Install]
WantedBy=multi-user.target
EOF

# Create environment files
echo "📝 Creating environment files..."
sudo tee /etc/quiver/gateway.env > /dev/null << EOF
QUIVER_GATEWAY_PORT=$QUIVER_GATEWAY_PORT
QUIVER_ENABLE_AUTH=$QUIVER_ENABLE_AUTH
QUIVER_JWT_SECRET=$QUIVER_JWT_SECRET
QUIVER_API_KEY_PREFIX=$QUIVER_API_KEY_PREFIX
QUIVER_BOOTSTRAP=$QUIVER_BOOTSTRAP
ENABLE_TLS=$ENABLE_TLS
TLS_CERT_PATH=$TLS_CERT_PATH
TLS_KEY_PATH=$TLS_KEY_PATH
LOG_LEVEL=$LOG_LEVEL
EOF

sudo tee /etc/quiver/provider.env > /dev/null << EOF
QUIVER_PROVIDER_PORT=$QUIVER_PROVIDER_PORT
QUIVER_OLLAMA_URL=$QUIVER_OLLAMA_URL
QUIVER_BOOTSTRAP=$QUIVER_BOOTSTRAP
LOG_LEVEL=$LOG_LEVEL
EOF

# Set permissions
sudo chmod 600 /etc/quiver/*.env

# Create user if not exists
if ! id -u quiver > /dev/null 2>&1; then
    echo "👤 Creating quiver user..."
    sudo useradd -r -s /bin/false -d /var/lib/quiver quiver
fi

# Create directories
sudo mkdir -p /var/log/quiver /var/lib/quiver
sudo chown -R quiver:quiver /var/log/quiver /var/lib/quiver

# Install binaries
echo "📦 Installing binaries..."
sudo cp gateway/gateway /usr/local/bin/quiver-gateway
sudo cp provider/provider /usr/local/bin/quiver-provider
sudo cp aggregator/aggregator /usr/local/bin/quiver-aggregator
sudo chmod 755 /usr/local/bin/quiver-*

# Setup log rotation
sudo tee /etc/logrotate.d/quiver > /dev/null << EOF
/var/log/quiver/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0640 quiver quiver
    postrotate
        systemctl reload quiver-gateway quiver-provider 2>/dev/null || true
    endscript
}
EOF

# Enable and start services
echo "🚀 Starting services..."
sudo systemctl daemon-reload
sudo systemctl enable quiver-gateway quiver-provider
sudo systemctl start quiver-gateway quiver-provider

# Wait for services to start
sleep 5

# Check status
echo ""
echo "✅ Deployment complete!"
echo ""
echo "Service Status:"
sudo systemctl status quiver-gateway --no-pager | grep Active
sudo systemctl status quiver-provider --no-pager | grep Active
echo ""
echo "Logs:"
echo "  sudo journalctl -u quiver-gateway -f"
echo "  sudo journalctl -u quiver-provider -f"
echo ""
echo "Test endpoint:"
echo "  curl http://localhost:$QUIVER_GATEWAY_PORT/health"