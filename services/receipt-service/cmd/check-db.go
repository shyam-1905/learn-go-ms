package main

import (
	"context"
	"expense-tracker/receipt-service/internal/config"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func checkDatabases() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	// Connect to postgres database to list all databases
	defaultConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
	)

	pool, err := pgxpool.New(ctx, defaultConnString)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	fmt.Println("========================================")
	fmt.Println("Database Check - Receipt Service")
	fmt.Println("========================================")
	fmt.Printf("Host: %s:%d\n", cfg.DBHost, cfg.DBPort)
	fmt.Printf("User: %s\n", cfg.DBUser)
	fmt.Println("")

	// Check receipt_db
	fmt.Println("----------------------------------------")
	fmt.Println("Checking receipt_db...")
	receiptConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/receipt_db?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
	)

	receiptPool, err := pgxpool.New(ctx, receiptConnString)
	if err != nil {
		fmt.Printf("  ERROR: Cannot connect to receipt_db: %v\n", err)
	} else {
		defer receiptPool.Close()

		// Check if receipts table exists
		var tableExists bool
		err = receiptPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'receipts')").Scan(&tableExists)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if tableExists {
			fmt.Println("  OK: receipts table exists")

			// Get table columns
			rows, err := receiptPool.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'receipts' ORDER BY ordinal_position")
			if err == nil {
				fmt.Println("  Columns:")
				for rows.Next() {
					var colName, dataType string
					rows.Scan(&colName, &dataType)
					fmt.Printf("    - %s (%s)\n", colName, dataType)
				}
				rows.Close()
			}

			// Get indexes
			rows, err = receiptPool.Query(ctx, "SELECT indexname FROM pg_indexes WHERE tablename = 'receipts' ORDER BY indexname")
			if err == nil {
				fmt.Println("  Indexes:")
				for rows.Next() {
					var indexName string
					rows.Scan(&indexName)
					fmt.Printf("    - %s\n", indexName)
				}
				rows.Close()
			}

			// Count receipts
			var receiptCount int
			receiptPool.QueryRow(ctx, "SELECT COUNT(*) FROM receipts WHERE deleted_at IS NULL").Scan(&receiptCount)
			fmt.Printf("  Active receipts: %d\n", receiptCount)
		} else {
			fmt.Println("  WARNING: receipts table does not exist")
		}
	}

	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("Database check completed!")
	fmt.Println("========================================")
}
