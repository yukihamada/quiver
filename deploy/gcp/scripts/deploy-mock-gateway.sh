#!/bin/bash
set -e

# Deploy mock gateway to all gateway instances
GATEWAY_IPS=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(networkInterfaces[0].accessConfigs[0].natIP)")

for IP in $GATEWAY_IPS; do
    echo "Deploying to $IP..."
    
    # Create mock gateway script
    gcloud compute ssh root@quiver-gateway-${IP##*.} --zone=asia-northeast1-a --command="cat > /tmp/deploy-mock.sh << 'EOF'
#!/bin/bash
set -e

# Stop existing service
systemctl stop quiver-gateway || true

# Create mock gateway
cd /opt/quiver/gateway
mkdir -p cmd/mock-gateway

# Create main.go
cat > cmd/mock-gateway/main.go << 'GOEOF'
package main

import (
    \"fmt\"
    \"log\"
    \"net/http\"
    \"encoding/json\"
    \"time\"
)

type Request struct {
    Prompt string \`json:\"prompt\"\`
    Model  string \`json:\"model\"\`
    Token  string \`json:\"token\"\`
}

type Response struct {
    Completion string \`json:\"completion\"\`
    Model      string \`json:\"model\"\`
    Receipt    Receipt \`json:\"receipt\"\`
}

type Receipt struct {
    Receipt   InnerReceipt \`json:\"receipt\"\`
    Signature string      \`json:\"signature\"\`
}

type InnerReceipt struct {
    ProviderPK string \`json:\"provider_pk\"\`
    Timestamp  int64  \`json:\"timestamp\"\`
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
    // CORS
    w.Header().Set(\"Access-Control-Allow-Origin\", \"*\")
    w.Header().Set(\"Access-Control-Allow-Methods\", \"POST, OPTIONS\")
    w.Header().Set(\"Access-Control-Allow-Headers\", \"Content-Type\")
    
    if r.Method == \"OPTIONS\" {
        w.WriteHeader(http.StatusNoContent)
        return
    }
    
    if r.Method != \"POST\" {
        http.Error(w, \"Method not allowed\", http.StatusMethodNotAllowed)
        return
    }
    
    var req Request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, \"Invalid request\", http.StatusBadRequest)
        return
    }
    
    if req.Prompt == \"\" {
        http.Error(w, \"Prompt is required\", http.StatusBadRequest)
        return
    }
    
    if req.Model == \"\" {
        req.Model = \"llama3.2:3b\"
    }
    
    resp := Response{
        Completion: fmt.Sprintf(\"QUIVer P2Pネットワークからの応答です。プロンプト『%s』を受け取りました。これはモックレスポンスですが、実際のP2Pネットワークでは分散型AIモデルが応答を生成します。\", req.Prompt),
        Model:      req.Model,
        Receipt: Receipt{
            Receipt: InnerReceipt{
                ProviderPK: \"12D3KooWMockGateway\",
                Timestamp:  time.Now().Unix(),
            },
            Signature: \"0xmocksignature123\",
        },
    }
    
    w.Header().Set(\"Content-Type\", \"application/json\")
    json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set(\"Content-Type\", \"application/json\")
    json.NewEncoder(w).Encode(map[string]interface{}{
        \"status\": \"healthy\",
        \"timestamp\": time.Now().Unix(),
    })
}

func main() {
    http.HandleFunc(\"/generate\", handleGenerate)
    http.HandleFunc(\"/health\", handleHealth)
    
    fmt.Println(\"Mock gateway starting on :8080\")
    log.Fatal(http.ListenAndServe(\":8080\", nil))
}
GOEOF

# Build mock gateway
go build -o /usr/local/bin/quiver-gateway-mock ./cmd/mock-gateway

# Update systemd service
cat > /etc/systemd/system/quiver-gateway.service << 'SVCEOF'
[Unit]
Description=QUIVer Mock Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/quiver-gateway-mock
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
SVCEOF

# Restart service
systemctl daemon-reload
systemctl enable quiver-gateway
systemctl start quiver-gateway

echo \"Mock gateway deployed successfully\"
EOF"
    
    # Execute deployment
    gcloud compute ssh root@quiver-gateway-${IP##*.} --zone=asia-northeast1-a --command="bash /tmp/deploy-mock.sh"
done

echo "Deployment complete!"