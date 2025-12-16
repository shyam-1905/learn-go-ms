//go:build !migrate

package main

import (
	"context"
	"expense-tracker/auth-service/internal/config"
	"expense-tracker/auth-service/internal/handler"
	"expense-tracker/auth-service/internal/middleware"
	"expense-tracker/auth-service/internal/repository"
	"expense-tracker/auth-service/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to PostgreSQL database
	// pgxpool.New creates a connection pool
	// A pool manages multiple connections efficiently
	log.Println("Connecting to database...")
	dbPool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close() // Close the pool when main exits

	// Test the connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established!")

	// Initialize repository (data access layer)
	userRepo := repository.NewPostgresUserRepository(dbPool)

	// Initialize JWT service
	jwtService := service.NewJWTService(cfg.JWTSecret, cfg.JWTExpiration)

	// Initialize auth service (business logic layer)
	authService := service.NewAuthService(userRepo, jwtService)

	// Initialize event publisher (optional - for notifications)
	if cfg.AuthEventsTopicARN != "" && cfg.AWSAccessKeyID != "" && cfg.AWSSecretKey != "" {
		log.Println("Initializing event publisher...")
		eventPublisher, err := service.NewEventPublisher(
			cfg.AWSRegion,
			cfg.AWSAccessKeyID,
			cfg.AWSSecretKey,
			cfg.AuthEventsTopicARN,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize event publisher: %v (events will not be published)", err)
		} else {
			authService.SetEventPublisher(eventPublisher)
			log.Println("Event publisher initialized!")
		}
	}

	// Initialize handlers (HTTP layer)
	authHandler := handler.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService) // For protected routes
	loggingMiddleware := middleware.NewLoggingMiddleware()     // For request/response logging

	// Setup HTTP router
	// gorilla/mux is a powerful HTTP router for Go
	router := mux.NewRouter()

	// Add logging middleware FIRST (so it logs all requests)
	// Middleware is executed in the order it's added
	router.Use(loggingMiddleware.LogRequest)

	// Add CORS middleware (allows frontend to call the API)
	// In production, restrict this to your frontend domain!
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Store authMiddleware for later use (we'll use it for protected routes)
	_ = authMiddleware

	// Define routes
	// Public routes (no authentication required)
	router.HandleFunc("/health", authHandler.Health).Methods("GET")
	router.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/auth/validate", authHandler.Validate).Methods("GET")

	// Protected routes (require authentication)
	// Example: router.HandleFunc("/auth/profile", authMiddleware.RequireAuth(getProfile)).Methods("GET")

	// Create HTTP server
	// http.Server is Go's built-in HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second, // Max time to read request
		WriteTimeout: 15 * time.Second, // Max time to write response
		IdleTimeout:  60 * time.Second, // Max time for idle connections
	}

	// Start server in a goroutine
	// A goroutine is a lightweight thread - it runs concurrently
	// This allows us to handle shutdown signals while the server runs
	go func() {
		log.Printf("Server starting on port %s...", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	// This is called "graceful shutdown" - we give the server time to finish requests
	log.Println("Server is running. Press Ctrl+C to stop.")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register for interrupt and terminate signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	// Give the server 30 seconds to finish handling requests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
