#!/bin/bash
set -e

# Deploy mock gateway to all gateway instances
echo "Getting gateway instances..."

# Get instance names directly
INSTANCES=$(gcloud compute instances list --filter="name:quiver-gateway-*" --format="value(name)")

for INSTANCE in $INSTANCES; do
    echo "Deploying to $INSTANCE..."
    
    # Create deployment script on the instance
    gcloud compute ssh root@$INSTANCE --zone=asia-northeast1-a --command="
# Stop existing service
systemctl stop quiver-gateway || true

# Create mock gateway binary directly
cat > /usr/local/bin/quiver-gateway-mock << 'EOF'
#!/usr/bin/env python3
import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler

class MockHandler(BaseHTTPRequestHandler):
    def do_OPTIONS(self):
        self.send_response(204)
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'POST, GET, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        self.end_headers()
    
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            response = {'status': 'healthy', 'timestamp': int(time.time())}
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_error(404)
    
    def do_POST(self):
        if self.path == '/generate':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            try:
                request = json.loads(post_data.decode())
                prompt = request.get('prompt', '')
                model = request.get('model', 'llama3.2:3b')
                
                if not prompt:
                    self.send_response(400)
                    self.send_header('Content-Type', 'application/json')
                    self.send_header('Access-Control-Allow-Origin', '*')
                    self.end_headers()
                    self.wfile.write(json.dumps({'error': 'Prompt is required'}).encode())
                    return
                
                response = {
                    'completion': f'QUIVer P2Pネットワークからの応答です。プロンプト「{prompt}」を受け取りました。これはモックレスポンスですが、実際のP2Pネットワークでは分散型AIモデルが応答を生成します。',
                    'model': model,
                    'receipt': {
                        'receipt': {
                            'provider_pk': '12D3KooWMockGateway',
                            'timestamp': int(time.time())
                        },
                        'signature': '0xmocksignature123'
                    }
                }
                
                self.send_response(200)
                self.send_header('Content-Type', 'application/json')
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(json.dumps(response).encode())
            
            except Exception as e:
                self.send_response(400)
                self.send_header('Content-Type', 'application/json')
                self.send_header('Access-Control-Allow-Origin', '*')
                self.end_headers()
                self.wfile.write(json.dumps({'error': 'Invalid request'}).encode())
        else:
            self.send_error(404)

if __name__ == '__main__':
    server = HTTPServer(('0.0.0.0', 8080), MockHandler)
    print('Mock gateway starting on :8080')
    server.serve_forever()
EOF

chmod +x /usr/local/bin/quiver-gateway-mock

# Update systemd service
cat > /etc/systemd/system/quiver-gateway.service << 'SVCEOF'
[Unit]
Description=QUIVer Mock Gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
ExecStart=/usr/bin/python3 /usr/local/bin/quiver-gateway-mock
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

# Check status
systemctl status quiver-gateway --no-pager || true
"
    
    echo "Deployed to $INSTANCE"
done

echo "Testing endpoints..."
sleep 5

# Test each gateway
for INSTANCE in $INSTANCES; do
    IP=$(gcloud compute instances describe $INSTANCE --zone=asia-northeast1-a --format="value(networkInterfaces[0].accessConfigs[0].natIP)")
    echo -n "Testing $INSTANCE ($IP): "
    curl -s http://$IP:8080/health | jq -r '.status' || echo "Failed"
done

echo "Deployment complete!"