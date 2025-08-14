#!/bin/bash

# QUIVer Demo Setup Script
# This script sets up QUIVer for inference from Mac and iPhone

set -e

echo "🚀 QUIVer P2P QUIC Provider Demo Setup"
echo "======================================"

# Check prerequisites
echo "📋 Checking prerequisites..."
if ! command -v ollama &> /dev/null; then
    echo "❌ Ollama not found. Please install it first: https://ollama.ai"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go not found. Please install Go 1.21+ first"
    exit 1
fi

# Create directories for keys and data
echo "📁 Creating directories..."
mkdir -p data/provider data/gateway data/aggregator

# Generate provider key if not exists
if [ ! -f data/provider/key.pem ]; then
    echo "🔐 Generating provider Ed25519 key..."
    openssl genpkey -algorithm ed25519 -out data/provider/key.pem
fi

# Start Ollama if not running
echo "🤖 Ensuring Ollama is running..."
if ! pgrep -x "ollama" > /dev/null; then
    ollama serve &
    sleep 2
fi

# Pull a model if not present
echo "📦 Ensuring Ollama model is available..."
ollama pull llama3.2:3b 2>/dev/null || true

# Create simple HTTP gateway for mobile access
cat > gateway_simple.go << 'EOF'
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
)

type Request struct {
    Prompt    string `json:"prompt"`
    Model     string `json:"model,omitempty"`
    MaxTokens int    `json:"max_tokens,omitempty"`
}

type Response struct {
    Completion string `json:"completion"`
    Model      string `json:"model"`
}

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var req Request
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        if req.Model == "" {
            req.Model = "llama3.2:3b"
        }

        // Call Ollama API
        ollamaReq := map[string]interface{}{
            "model":  req.Model,
            "prompt": req.Prompt,
            "stream": false,
            "options": map[string]interface{}{
                "temperature": 0,
                "seed":        42,
            },
        }

        reqBody, _ := json.Marshal(ollamaReq)
        resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(reqBody))
        if err != nil {
            http.Error(w, "Failed to call Ollama", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        body, _ := io.ReadAll(resp.Body)
        var ollamaResp map[string]interface{}
        json.Unmarshal(body, &ollamaResp)

        response := Response{
            Completion: ollamaResp["response"].(string),
            Model:      req.Model,
        }

        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            return
        }

        json.NewEncoder(w).Encode(response)
    })

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
    })

    fmt.Printf("🌐 Gateway listening on http://0.0.0.0:%s\n", port)
    http.ListenAndServe(":"+port, nil)
}
EOF

# Build and run the simple gateway
echo "🔨 Building simple gateway..."
go build -o bin/gateway_simple gateway_simple.go

# Get local IP for network access
LOCAL_IP=$(ipconfig getifaddr en0 || echo "localhost")

echo ""
echo "✅ Setup complete!"
echo ""
echo "📱 To start the service for Mac and iPhone access:"
echo ""
echo "   ./bin/gateway_simple"
echo ""
echo "📲 Access from devices on the same network:"
echo ""
echo "   Mac:    http://localhost:8080/generate"
echo "   iPhone: http://$LOCAL_IP:8080/generate"
echo ""
echo "📝 Example usage:"
echo ""
echo '   curl -X POST http://localhost:8080/generate \'
echo '     -H "Content-Type: application/json" \'
echo '     -d "{\"prompt\": \"What is the capital of France?\"}"'
echo ""
echo "🔧 For iOS, you can use shortcuts or any HTTP client app"
echo ""

# Create iOS Shortcut instructions
cat > iOS_SETUP.md << EOF
# iOS Setup Instructions

## Using Shortcuts App

1. Open Shortcuts app on iPhone
2. Create new shortcut
3. Add action: "Get Contents of URL"
4. Set URL: http://$LOCAL_IP:8080/generate
5. Set Method: POST
6. Add Headers:
   - Content-Type: application/json
7. Set Body: JSON with format:
   \`\`\`json
   {
     "prompt": "Your question here"
   }
   \`\`\`
8. Add action: "Get Dictionary from Input"
9. Add action: "Get Text from Input" (for completion field)
10. Save and run!

## Using HTTP Client Apps

Recommended apps:
- HTTP Bot
- RESTed
- Paw

Configure with:
- URL: http://$LOCAL_IP:8080/generate
- Method: POST
- Headers: Content-Type: application/json
- Body: {"prompt": "Your question"}
EOF

echo "📱 iOS setup instructions saved to iOS_SETUP.md"