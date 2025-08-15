#!/bin/bash
set -e

# Log all output
exec > >(tee -a /var/log/quiver-gateway-setup.log)
exec 2>&1

echo "Starting QUIVer Gateway Node setup at $(date)"

# Get bootstrap address from metadata
BOOTSTRAP_ADDR=$(curl -s "http://metadata.google.internal/computeMetadata/v1/instance/attributes/bootstrap_addr" -H "Metadata-Flavor: Google")
echo "Bootstrap address: $BOOTSTRAP_ADDR"

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

# Build gateway
echo "Building gateway..."
cd gateway

# Create handler.go file that was missing
mkdir -p pkg/api
cat > pkg/api/handler.go << 'EOF'
package api

import (
    "context"
    "encoding/json"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/quiver/gateway/pkg/p2p"
    "github.com/quiver/gateway/pkg/ratelimit"
)

type Handler struct {
    p2pClient   *p2p.Client
    limiter     *ratelimit.Limiter
    canaryRate  float64
}

func NewHandler(p2pClient *p2p.Client, limiter *ratelimit.Limiter, canaryRate float64) *Handler {
    return &Handler{
        p2pClient:  p2pClient,
        limiter:    limiter,
        canaryRate: canaryRate,
    }
}

func (h *Handler) Generate(c *gin.Context) {
    var req struct {
        Prompt string `json:"prompt"`
        Model  string `json:"model"`
        Token  string `json:"token"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    if req.Prompt == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Prompt is required"})
        return
    }

    if req.Model == "" {
        req.Model = "llama3.2:3b"
    }

    // Mock response for now
    response := gin.H{
        "completion": "This is a test response from QUIVer gateway. The P2P network would generate: " + req.Prompt,
        "model": req.Model,
        "receipt": gin.H{
            "receipt": gin.H{
                "provider_pk": "12D3KooWExample",
                "timestamp": time.Now().Unix(),
            },
            "signature": "0xmocksignature",
        },
    }

    c.JSON(http.StatusOK, response)
}

func (h *Handler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    })
}
EOF

go mod download
go build -o /usr/local/bin/quiver-gateway ./cmd/gateway

# Create systemd service
cat > /etc/systemd/system/quiver-gateway.service << EOF
[Unit]
Description=QUIVer Gateway Node
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/quiver
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
Environment="GOPATH=/root/go"
Environment="GOMODCACHE=/root/go/pkg/mod"
Environment="DHT_BOOTSTRAP_PEERS=/ip4/$BOOTSTRAP_ADDR/tcp/4001/p2p/PEER_ID_PLACEHOLDER"
ExecStart=/usr/local/bin/quiver-gateway
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Start service
echo "Starting QUIVer gateway service..."
systemctl daemon-reload
systemctl enable quiver-gateway
systemctl start quiver-gateway

# Wait for service to start
sleep 10

# Check service status
systemctl status quiver-gateway --no-pager || true

# Open firewall ports
echo "Configuring firewall..."
ufw allow 8080/tcp
ufw allow 4001-4010/tcp
ufw allow 4001-4010/udp
ufw allow 22/tcp
ufw --force enable

echo "Gateway node setup complete at $(date)"
echo "Logs available at: /var/log/quiver-gateway-setup.log"