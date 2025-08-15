package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/quiver/billing/pkg/auth"
	"github.com/quiver/billing/pkg/db"
	"github.com/quiver/billing/pkg/handlers"
	"github.com/quiver/billing/pkg/stripe"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	database, err := db.New(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize Stripe client
	stripeClient := stripe.NewClient(
		os.Getenv("STRIPE_API_KEY"),
		os.Getenv("STRIPE_WEBHOOK_SECRET"),
	)

	// Initialize JWT token manager
	tokenManager := auth.NewTokenManager(
		os.Getenv("JWT_SECRET"),
		24*time.Hour,
	)

	// Initialize handlers
	h := handlers.New(database, stripeClient, tokenManager)

	// Setup router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://quiver.network", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Public routes
	public := r.Group("/api/v1")
	{
		public.POST("/auth/signup", h.SignUp)
		public.POST("/auth/login", h.Login)
		public.GET("/plans", h.GetPlans)
		
		// Stripe webhooks
		public.POST("/webhooks/stripe", h.HandleStripeWebhook)
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(h.AuthMiddleware())
	{
		// User management
		protected.GET("/user", h.GetCurrentUser)
		protected.PUT("/user", h.UpdateUser)
		
		// API keys
		protected.POST("/api-keys", h.CreateAPIKey)
		protected.GET("/api-keys", h.ListAPIKeys)
		protected.DELETE("/api-keys/:id", h.DeleteAPIKey)
		
		// Subscription management
		protected.POST("/subscriptions", h.CreateSubscription)
		protected.PUT("/subscriptions", h.UpdateSubscription)
		protected.DELETE("/subscriptions", h.CancelSubscription)
		protected.GET("/subscriptions", h.GetSubscription)
		
		// Usage statistics
		protected.GET("/usage", h.GetUsage)
		protected.GET("/usage/current", h.GetCurrentUsage)
		
		// Billing
		protected.GET("/invoices", h.ListInvoices)
		protected.POST("/payment-methods", h.AddPaymentMethod)
		protected.GET("/payment-methods", h.ListPaymentMethods)
		protected.DELETE("/payment-methods/:id", h.DeletePaymentMethod)
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "billing",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Billing service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}