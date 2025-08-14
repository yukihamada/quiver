package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type NodeInfo struct {
	PeerID    string    `json:"peer_id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Location  string    `json:"location"`
	Capacity  float64   `json:"capacity"`
	LastSeen  time.Time `json:"last_seen"`
	Version   string    `json:"version"`
	Metrics   NodeMetrics `json:"metrics"`
}

type NodeMetrics struct {
	RequestsHandled int64   `json:"requests_handled"`
	TotalTokens     int64   `json:"total_tokens"`
	AverageLatency  float64 `json:"average_latency_ms"`
	Uptime          float64 `json:"uptime_seconds"`
}

type NetworkStats struct {
	NodeCount      int                   `json:"node_count"`
	OnlineNodes    int                   `json:"online_nodes"`
	Countries      int                   `json:"countries"`
	TotalCapacity  float64               `json:"total_capacity"`
	Throughput     float64               `json:"throughput"`
	LastUpdate     time.Time             `json:"last_update"`
	NetworkHealth  string                `json:"network_health"`
	NetworkType    string                `json:"network_type"`
	Growth24h      string                `json:"growth_24h"`
	Version        string                `json:"version"`
	Nodes          map[string]int        `json:"nodes"`
	Regions        map[string]int        `json:"regions"`
	NodeDetails    map[string]NodeInfo   `json:"node_details"`
	Performance    PerformanceMetrics    `json:"performance"`
}

type PerformanceMetrics struct {
	AverageResponseTime float64 `json:"average_response_time_ms"`
	RequestsPerMinute   int64   `json:"requests_per_minute"`
	TokensPerSecond     float64 `json:"tokens_per_second"`
	SuccessRate         float64 `json:"success_rate"`
}

type StatsServer struct {
	mu              sync.RWMutex
	nodes           map[string]NodeInfo
	stats           NetworkStats
	clients         map[*websocket.Conn]bool
	nodeHealthCheck map[string]string
	upgrader        websocket.Upgrader
}

func NewStatsServer() *StatsServer {
	return &StatsServer{
		nodes:   make(map[string]NodeInfo),
		clients: make(map[*websocket.Conn]bool),
		nodeHealthCheck: map[string]string{
			"http://localhost:8090/health": "local",
			"http://localhost:8091/health": "local",
			// Add GCP nodes when deployed
			// "http://34.146.32.216:8090/health": "gcp",
		},
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for GitHub Pages
			},
		},
	}
}

func (s *StatsServer) collectNodeInfo(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.updateNodeInfo()
		}
	}
}

func (s *StatsServer) updateNodeInfo() {
	client := &http.Client{Timeout: 5 * time.Second}
	activeNodes := make(map[string]NodeInfo)

	// Check each known node endpoint
	for endpoint, location := range s.nodeHealthCheck {
		resp, err := client.Get(endpoint)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var health map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			continue
		}

		peerID, _ := health["peer_id"].(string)
		if peerID == "" {
			continue
		}

		nodeType := "provider"
		if service, ok := health["service"].(string); ok {
			nodeType = service
		}

		uptime := float64(0)
		if u, ok := health["uptime_seconds"].(float64); ok {
			uptime = u
		}

		node := NodeInfo{
			PeerID:   peerID,
			Type:     nodeType,
			Status:   "healthy",
			Location: location,
			Capacity: 1.2, // Default capacity
			LastSeen: time.Now(),
			Version:  "1.0.0",
			Metrics: NodeMetrics{
				Uptime: uptime,
			},
		}

		activeNodes[peerID] = node
	}

	// Update stored nodes
	s.mu.Lock()
	s.nodes = activeNodes
	s.updateStats()
	s.mu.Unlock()

	// Broadcast to WebSocket clients
	s.broadcastStats()
}

func (s *StatsServer) updateStats() {
	nodeTypes := make(map[string]int)
	regions := make(map[string]int)
	totalCapacity := 0.0
	onlineCount := 0

	for _, node := range s.nodes {
		nodeTypes[node.Type]++
		regions[node.Location]++
		totalCapacity += node.Capacity
		if node.Status == "healthy" {
			onlineCount++
		}
	}

	s.stats = NetworkStats{
		NodeCount:     len(s.nodes),
		OnlineNodes:   onlineCount,
		Countries:     1, // Update based on actual geo data
		TotalCapacity: totalCapacity,
		Throughput:    float64(len(s.nodes)) * 150, // Estimate
		LastUpdate:    time.Now(),
		NetworkHealth: s.calculateNetworkHealth(),
		NetworkType:   "testnet",
		Growth24h:     "0%",
		Version:       "1.0.0",
		Nodes:         nodeTypes,
		Regions:       regions,
		NodeDetails:   s.nodes,
		Performance: PerformanceMetrics{
			AverageResponseTime: 2500,
			RequestsPerMinute:   int64(len(s.nodes) * 10),
			TokensPerSecond:     float64(len(s.nodes) * 25),
			SuccessRate:         98.5,
		},
	}
}

func (s *StatsServer) calculateNetworkHealth() string {
	if len(s.nodes) == 0 {
		return "offline"
	}
	if len(s.nodes) < 3 {
		return "degraded"
	}
	return "healthy"
}

func (s *StatsServer) broadcastStats() {
	s.mu.RLock()
	message, _ := json.Marshal(s.stats)
	s.mu.RUnlock()

	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			client.Close()
			delete(s.clients, client)
		}
	}
}

func (s *StatsServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	// Send initial stats
	s.mu.RLock()
	initialStats, _ := json.Marshal(s.stats)
	s.mu.RUnlock()
	conn.WriteMessage(websocket.TextMessage, initialStats)

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			break
		}
	}
}

func (s *StatsServer) handleStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.stats)
}

func (s *StatsServer) handleNodeRegister(w http.ResponseWriter, r *http.Request) {
	var node NodeInfo
	if err := json.NewDecoder(r.Body).Decode(&node); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	node.LastSeen = time.Now()
	s.nodes[node.PeerID] = node
	s.updateStats()
	s.mu.Unlock()

	s.broadcastStats()
	w.WriteHeader(http.StatusOK)
}

func main() {
	server := NewStatsServer()
	ctx := context.Background()

	// Start node info collector
	go server.collectNodeInfo(ctx)

	// Initial update
	server.updateNodeInfo()

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/api/stats", server.handleStats).Methods("GET")
	router.HandleFunc("/api/nodes/register", server.handleNodeRegister).Methods("POST")
	router.HandleFunc("/ws", server.handleWebSocket)

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(router)
	
	fmt.Println("Stats server starting on :8087")
	log.Fatal(http.ListenAndServe(":8087", handler))
}