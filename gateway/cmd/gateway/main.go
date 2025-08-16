package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/quiver/gateway/internal/config"
	"github.com/quiver/gateway/pkg/api"
	"github.com/quiver/gateway/pkg/auth"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/quiver/gateway/pkg/ratelimit"
)

func main() {
	cfg := config.DefaultConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p2pClient, err := p2p.NewClient(ctx, cfg.P2PListenAddr, cfg.DHTBootstrapPeers)
	if err != nil {
		log.Fatal("Failed to create P2P client:", err)
	}
	defer p2pClient.Close()

	limiter := ratelimit.NewLimiter(cfg.RateLimitPerToken)
	go limiter.CleanupOldLimiters()

	handler := api.NewHandler(p2pClient, limiter, cfg.CanaryRate)

	// Initialize authenticator
	authConfig := auth.AuthConfig{
		JWTSecret:    []byte(cfg.JWTSecret),
		APIKeyPrefix: cfg.APIKeyPrefix,
		EnableAuth:   cfg.EnableAuth,
	}
	authenticator := auth.NewAuthenticator(authConfig)
	rateLimiter := auth.NewRateLimiter()

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

	// Public endpoints
	router.GET("/health", handler.Health)
	router.GET("/stats", handler.StatsHandler)
	router.OPTIONS("/generate", func(c *gin.Context) {
		c.Status(204)
	})
	router.OPTIONS("/stats", func(c *gin.Context) {
		c.Status(204)
	})

	// Protected endpoints
	protected := router.Group("")
	if cfg.EnableAuth {
		protected.Use(authenticator.AuthMiddleware())
		protected.Use(rateLimiter.RateLimitMiddleware())
	}
	protected.POST("/generate", handler.Generate)

	fmt.Printf("Gateway started on port %s\n", cfg.Port)

	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}
