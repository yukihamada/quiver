package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/quiver/aggregator/internal/config"
	"github.com/quiver/aggregator/pkg/api"
	"github.com/quiver/aggregator/pkg/epoch"
	"github.com/quiver/aggregator/pkg/storage"
)

func main() {
	cfg := config.DefaultConfig()

	store := storage.NewStore()
	epochManager := epoch.NewManager()
	handler := api.NewHandler(store, epochManager)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/commit", handler.Commit)
	router.POST("/claim", handler.Claim)
	router.GET("/state", handler.GetState)
	router.GET("/health", handler.Health)

	fmt.Printf("Aggregator started on port %s\n", cfg.Port)

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
