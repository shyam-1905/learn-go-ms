//go:build !migrate

package main

import (
	"context"
	"expense-tracker/receipt-service/internal/config"
	"expense-tracker/receipt-service/internal/handler"
	"expense-tracker/receipt-service/internal/middleware"
	"expense-tracker/receipt-service/internal/repository"
	"expense-tracker/receipt-service/internal/service"
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
	log.Println("Connecting to database...")
	dbPool, err := pgxpool.New(context.Background(), cfg.GetDatabaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Test the connection
	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection established!")

	// Initialize AWS S3 service
	log.Println("Initializing AWS S3 service...")
	s3Service, err := service.NewS3Service(
		cfg.AWSRegion,
		cfg.AWSAccessKeyID,
		cfg.AWSSecretKey,
		cfg.S3BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to initialize S3 service: %v", err)
	}
	log.Println("S3 service initialized!")

	// Initialize repository (data access layer)
	receiptRepo := repository.NewPostgresReceiptRepository(dbPool)

	// Initialize auth client (for token validation via auth-service)
	authClient := service.NewAuthClient(cfg.AuthServiceURL)

	// Initialize receipt service (business logic layer)
	receiptService := service.NewReceiptService(receiptRepo, s3Service)

	// Initialize event publisher (optional - for notifications)
	if cfg.ReceiptEventsTopicARN != "" && cfg.AWSAccessKeyID != "" && cfg.AWSSecretKey != "" {
		log.Println("Initializing event publisher...")
		eventPublisher, err := service.NewEventPublisher(
			cfg.AWSRegion,
			cfg.AWSAccessKeyID,
			cfg.AWSSecretKey,
			cfg.ReceiptEventsTopicARN,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize event publisher: %v (events will not be published)", err)
		} else {
			receiptService.SetEventPublisher(eventPublisher)
			log.Println("Event publisher initialized!")
		}
	}

	// Initialize handlers (HTTP layer)
	receiptHandler := handler.NewReceiptHandler(receiptService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authClient) // For token validation via auth-service
	loggingMiddleware := middleware.NewLoggingMiddleware()     // For request/response logging

	// Setup HTTP router
	router := mux.NewRouter()

	// Add logging middleware FIRST (so it logs all requests)
	router.Use(loggingMiddleware.LogRequest)

	// Add CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Define routes
	// Public routes (no authentication required)
	router.HandleFunc("/health", receiptHandler.Health).Methods("GET")

	// Protected routes (require JWT authentication)
	// All receipt endpoints require authentication
	// IMPORTANT: More specific routes (like /receipts/{id}/link) must be defined
	// BEFORE routes with path variables (like /receipts/{id}) to avoid route conflicts
	router.HandleFunc("/receipts", authMiddleware.RequireAuth(receiptHandler.UploadReceipt)).Methods("POST")
	router.HandleFunc("/receipts", authMiddleware.RequireAuth(receiptHandler.ListReceipts)).Methods("GET")
	router.HandleFunc("/receipts/{id}/link", authMiddleware.RequireAuth(receiptHandler.LinkReceipt)).Methods("PUT")
	router.HandleFunc("/receipts/{id}", authMiddleware.RequireAuth(receiptHandler.GetReceipt)).Methods("GET")
	router.HandleFunc("/receipts/{id}", authMiddleware.RequireAuth(receiptHandler.DeleteReceipt)).Methods("DELETE")

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	log.Println("Server is running. Press Ctrl+C to stop.")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register for interrupt and terminate signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
