//go:build !migrate

package main

import (
	"context"
	"expense-tracker/expense-service/internal/config"
	"expense-tracker/expense-service/internal/handler"
	"expense-tracker/expense-service/internal/middleware"
	"expense-tracker/expense-service/internal/repository"
	"expense-tracker/expense-service/internal/service"
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

	// Initialize repository (data access layer)
	expenseRepo := repository.NewPostgresExpenseRepository(dbPool)

	// Initialize auth client (for token validation via auth-service)
	authClient := service.NewAuthClient(cfg.AuthServiceURL)

	// Initialize expense service (business logic layer)
	expenseService := service.NewExpenseService(expenseRepo)

	// Initialize event publisher (optional - for notifications)
	if cfg.ExpenseEventsTopicARN != "" && cfg.AWSAccessKeyID != "" && cfg.AWSSecretKey != "" {
		log.Println("Initializing event publisher...")
		log.Printf("  Topic ARN: %s", cfg.ExpenseEventsTopicARN)
		log.Printf("  AWS Region: %s", cfg.AWSRegion)
		eventPublisher, err := service.NewEventPublisher(
			cfg.AWSRegion,
			cfg.AWSAccessKeyID,
			cfg.AWSSecretKey,
			cfg.ExpenseEventsTopicARN,
		)
		if err != nil {
			log.Printf("ERROR: Failed to initialize event publisher: %v (events will not be published)", err)
		} else {
			expenseService.SetEventPublisher(eventPublisher)
			log.Println("✓ Event publisher initialized successfully!")
		}
	} else {
		log.Println("WARNING: Event publisher not initialized - missing configuration:")
		if cfg.ExpenseEventsTopicARN == "" {
			log.Println("  - EXPENSE_EVENTS_TOPIC_ARN is not set")
		}
		if cfg.AWSAccessKeyID == "" {
			log.Println("  - AWS_ACCESS_KEY_ID is not set")
		}
		if cfg.AWSSecretKey == "" {
			log.Println("  - AWS_SECRET_ACCESS_KEY is not set")
		}
		log.Println("  Events will not be published until configuration is complete.")
	}

	// Initialize handlers (HTTP layer)
	expenseHandler := handler.NewExpenseHandler(expenseService)

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
	router.HandleFunc("/health", expenseHandler.Health).Methods("GET")

	// Protected routes (require JWT authentication)
	// All expense endpoints require authentication
	// IMPORTANT: More specific routes (like /expenses/summary) must be defined
	// BEFORE routes with path variables (like /expenses/{id}) to avoid route conflicts
	router.HandleFunc("/expenses", authMiddleware.RequireAuth(expenseHandler.CreateExpense)).Methods("POST")
	router.HandleFunc("/expenses", authMiddleware.RequireAuth(expenseHandler.ListExpenses)).Methods("GET")
	router.HandleFunc("/expenses/summary", authMiddleware.RequireAuth(expenseHandler.GetSummary)).Methods("GET")
	router.HandleFunc("/expenses/{id}", authMiddleware.RequireAuth(expenseHandler.GetExpense)).Methods("GET")
	router.HandleFunc("/expenses/{id}", authMiddleware.RequireAuth(expenseHandler.UpdateExpense)).Methods("PUT")
	router.HandleFunc("/expenses/{id}", authMiddleware.RequireAuth(expenseHandler.DeleteExpense)).Methods("DELETE")

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
