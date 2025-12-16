package main

import (
	"context"
	"expense-tracker/notification-service/internal/config"
	"expense-tracker/notification-service/internal/handler"
	"expense-tracker/notification-service/internal/middleware"
	"expense-tracker/notification-service/internal/model"
	"expense-tracker/notification-service/internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize SNS publisher for sending emails
	log.Println("Initializing SNS publisher...")
	snsPublisher, err := service.NewSNSPublisher(
		cfg.AWSRegion,
		cfg.AWSAccessKeyID,
		cfg.AWSSecretKey,
		cfg.NotificationEmailTopicARN,
	)
	if err != nil {
		log.Fatalf("Failed to initialize SNS publisher: %v", err)
	}
	log.Println("SNS publisher initialized!")

	// Initialize template service
	log.Println("Initializing template service...")
	// Determine templates directory path
	// Try multiple locations to support different execution contexts
	var templatesDir string
	possiblePaths := []string{
		filepath.Join(".", "templates"),                                // Current directory
		filepath.Join("services", "notification-service", "templates"), // From workspace root
		filepath.Join("notification-service", "templates"),             // Alternative
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			templatesDir = path
			break
		}
	}

	// If none found, use current directory (will create templates there)
	if templatesDir == "" {
		templatesDir = filepath.Join(".", "templates")
		// Create directory if it doesn't exist
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			log.Printf("Warning: Could not create templates directory: %v", err)
		}
	}

	log.Printf("Using templates directory: %s", templatesDir)
	templateService, err := service.NewTemplateService(templatesDir)
	if err != nil {
		log.Fatalf("Failed to initialize template service: %v", err)
	}
	log.Println("Template service initialized!")

	// Initialize notification service
	notificationService := service.NewNotificationService(snsPublisher, templateService)

	// Initialize SQS consumers for each queue
	log.Println("Initializing SQS consumers...")
	sqsConsumer, err := service.NewSQSConsumer(
		cfg.AWSRegion,
		cfg.AWSAccessKeyID,
		cfg.AWSSecretKey,
	)
	if err != nil {
		log.Fatalf("Failed to initialize SQS consumer: %v", err)
	}
	log.Println("SQS consumer initialized!")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming from expense events queue
	if cfg.ExpenseEventsQueueURL != "" {
		go func() {
			log.Printf("Starting consumer for expense events queue: %s", cfg.ExpenseEventsQueueURL)
			if err := sqsConsumer.ConsumeMessages(ctx, cfg.ExpenseEventsQueueURL, func(ctx context.Context, event *model.Event) error {
				return notificationService.ProcessEvent(ctx, event)
			}); err != nil {
				log.Printf("Error consuming expense events: %v", err)
			}
		}()
	}

	// Start consuming from receipt events queue
	if cfg.ReceiptEventsQueueURL != "" {
		go func() {
			log.Printf("Starting consumer for receipt events queue: %s", cfg.ReceiptEventsQueueURL)
			if err := sqsConsumer.ConsumeMessages(ctx, cfg.ReceiptEventsQueueURL, func(ctx context.Context, event *model.Event) error {
				return notificationService.ProcessEvent(ctx, event)
			}); err != nil {
				log.Printf("Error consuming receipt events: %v", err)
			}
		}()
	}

	// Start consuming from auth events queue
	if cfg.AuthEventsQueueURL != "" {
		go func() {
			log.Printf("Starting consumer for auth events queue: %s", cfg.AuthEventsQueueURL)
			if err := sqsConsumer.ConsumeMessages(ctx, cfg.AuthEventsQueueURL, func(ctx context.Context, event *model.Event) error {
				return notificationService.ProcessEvent(ctx, event)
			}); err != nil {
				log.Printf("Error consuming auth events: %v", err)
			}
		}()
	}

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()

	// Initialize middleware
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Setup HTTP router
	router := mux.NewRouter()

	// Add logging middleware
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
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")

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
	log.Println("Notification service is running. Press Ctrl+C to stop.")
	log.Println("Consuming events from SQS queues...")

	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	// Register for interrupt and terminate signals
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal
	<-sigChan

	log.Println("Shutting down notification service...")

	// Cancel context to stop SQS consumers
	cancel()

	// Create a context with timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Notification service stopped")
}
