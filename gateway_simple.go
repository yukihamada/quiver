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
