package main

import (
    "fmt"
    "log"
    
    "github.com/gin-gonic/gin"
    "github.com/quiver/gateway/pkg/api"
)

func main() {
    gin.SetMode(gin.ReleaseMode)
    router := gin.New()
    router.Use(gin.Recovery())
    
    // CORS middleware
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    })
    
    handler := api.NewMockHandler()
    
    router.POST("/generate", handler.Generate)
    router.GET("/health", handler.Health)
    router.OPTIONS("/generate", func(c *gin.Context) {
        c.Status(204)
    })
    
    port := "8080"
    fmt.Printf("Mock Gateway started on port %s\n", port)
    
    if err := router.Run(":" + port); err != nil {
        log.Fatal("Failed to start server:", err)
    }
}