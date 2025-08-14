package api

import (
    "encoding/json"
    "net/http"
    "github.com/quiver/provider/pkg/p2p"
)

// Server handles API requests for the provider
type Server struct {
    node      *p2p.DecentralizedNode
    ollamaURL string
}

// NewServer creates a new API server
func NewServer(node *p2p.DecentralizedNode, ollamaURL string) *Server {
    return &Server{
        node:      node,
        ollamaURL: ollamaURL,
    }
}

// Start starts the API server
func (s *Server) Start(addr string) error {
    mux := http.NewServeMux()
    
    // API endpoints
    mux.HandleFunc("/api/inference", s.handleInference)
    mux.HandleFunc("/api/stats", s.handleStats)
    mux.HandleFunc("/api/health", s.handleHealth)
    
    return http.ListenAndServe(addr, mux)
}

func (s *Server) handleInference(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req struct {
        Model      string `json:"model"`
        Prompt     string `json:"prompt"`
        MaxTokens  int    `json:"max_tokens"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Send inference request through P2P network
    result, err := s.node.SendInferenceRequest(req.Model, req.Prompt, req.MaxTokens)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "result": result,
    })
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
    stats := s.node.GetNetworkStats()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "healthy",
        "node_id": s.node.GetPeerID(),
    })
}