package api

import (
    "encoding/json"
    "net/http"
    "os/exec"
    "runtime"
    "sync"
    "time"
)

type GUIServer struct {
    mu       sync.RWMutex
    stats    *Stats
    isRunning bool
}

type Stats struct {
    Earnings     float64   `json:"earnings"`
    Requests     int       `json:"requests"`
    Uptime       string    `json:"uptime"`
    CPUUsage     float64   `json:"cpu_usage"`
    MemoryUsage  float64   `json:"memory_usage"`
    NetworkStatus string   `json:"network_status"`
    Model        string    `json:"model"`
    StartTime    time.Time `json:"start_time"`
}

func NewGUIServer() *GUIServer {
    return &GUIServer{
        stats: &Stats{
            StartTime: time.Now(),
            Model: "llama3.2",
            NetworkStatus: "Connecting...",
        },
    }
}

func (g *GUIServer) Start(addr string) error {
    mux := http.NewServeMux()
    
    // Enable CORS
    handler := func(h http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", "*")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
            
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
            
            h(w, r)
        }
    }
    
    mux.HandleFunc("/stats", handler(g.handleStats))
    mux.HandleFunc("/start", handler(g.handleStart))
    mux.HandleFunc("/stop", handler(g.handleStop))
    mux.HandleFunc("/setup", handler(g.handleSetup))
    
    return http.ListenAndServe(addr, mux)
}

func (g *GUIServer) handleStats(w http.ResponseWriter, r *http.Request) {
    g.mu.RLock()
    defer g.mu.RUnlock()
    
    // Update uptime
    g.stats.Uptime = time.Since(g.stats.StartTime).String()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(g.stats)
}

func (g *GUIServer) handleStart(w http.ResponseWriter, r *http.Request) {
    g.mu.Lock()
    g.isRunning = true
    g.stats.NetworkStatus = "P2P Connected"
    g.mu.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (g *GUIServer) handleStop(w http.ResponseWriter, r *http.Request) {
    g.mu.Lock()
    g.isRunning = false
    g.stats.NetworkStatus = "Offline"
    g.mu.Unlock()
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (g *GUIServer) handleSetup(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Action string `json:"action"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    switch req.Action {
    case "download-model":
        // Run ollama pull command
        cmd := exec.Command("ollama", "pull", "llama3.2:3b")
        if err := cmd.Run(); err != nil {
            http.Error(w, "Failed to download model", http.StatusInternalServerError)
            return
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (g *GUIServer) UpdateStats(earnings float64, requests int) {
    g.mu.Lock()
    defer g.mu.Unlock()
    
    g.stats.Earnings = earnings
    g.stats.Requests = requests
}

func (g *GUIServer) OpenBrowser(url string) error {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "darwin":
        cmd = "open"
        args = []string{url}
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start", url}
    default:
        cmd = "xdg-open"
        args = []string{url}
    }

    return exec.Command(cmd, args...).Start()
}