package api

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

// MockHandler provides mock responses for testing
type MockHandler struct{}

// NewMockHandler creates a new mock handler
func NewMockHandler() *MockHandler {
    return &MockHandler{}
}

// Generate handles generate requests with mock responses
func (h *MockHandler) Generate(c *gin.Context) {
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
    
    // Mock response
    response := gin.H{
        "completion": "こんにちは！QUIVer P2Pネットワークからのテスト応答です。「" + req.Prompt + "」というプロンプトを受け取りました。",
        "model": req.Model,
        "receipt": gin.H{
            "receipt": gin.H{
                "provider_pk": "12D3KooWMockProvider",
                "timestamp": time.Now().Unix(),
            },
            "signature": "0xmocksignature",
        },
    }
    
    c.JSON(http.StatusOK, response)
}

// Health handles health check requests
func (h *MockHandler) Health(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "version": "mock-v1.0",
    })
}