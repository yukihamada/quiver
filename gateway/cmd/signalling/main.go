package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/quiver/gateway/pkg/signalling"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8444"
	}

	// Create signalling server
	server := signalling.NewSignallingServer()

	// Set up HTTP server
	http.HandleFunc("/signal", server.ServeHTTP)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"signalling"}`))
	})

	// CORS headers for all requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	fmt.Printf("Signalling server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}