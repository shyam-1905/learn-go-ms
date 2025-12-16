package main

import (
	"context"
	"expense-tracker/expense-service/internal/config"
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
	fmt.Println("Database Check")
	fmt.Println("========================================")
	fmt.Printf("Host: %s:%d\n", cfg.DBHost, cfg.DBPort)
	fmt.Printf("User: %s\n", cfg.DBUser)
	fmt.Println("")

	// List all databases
	fmt.Println("Available databases:")
	rows, err := pool.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		log.Fatalf("Failed to query databases: %v", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}
		databases = append(databases, dbName)
		fmt.Printf("  - %s\n", dbName)
	}
	fmt.Println("")

	// Check auth_db
	fmt.Println("----------------------------------------")
	fmt.Println("Checking auth_db...")
	authConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/auth_db?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
	)

	authPool, err := pgxpool.New(ctx, authConnString)
	if err != nil {
		fmt.Printf("  ERROR: Cannot connect to auth_db: %v\n", err)
	} else {
		defer authPool.Close()

		// Check if users table exists
		var tableExists bool
		err = authPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if tableExists {
			fmt.Println("  OK: users table exists")

			// Get table columns
			rows, err := authPool.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'users' ORDER BY ordinal_position")
			if err == nil {
				fmt.Println("  Columns:")
				for rows.Next() {
					var colName, dataType string
					rows.Scan(&colName, &dataType)
					fmt.Printf("    - %s (%s)\n", colName, dataType)
				}
				rows.Close()
			}

			// Count users
			var userCount int
			authPool.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&userCount)
			fmt.Printf("  Active users: %d\n", userCount)
		} else {
			fmt.Println("  WARNING: users table does not exist")
		}
	}
	fmt.Println("")

	// Check expense_db
	fmt.Println("----------------------------------------")
	fmt.Println("Checking expense_db...")
	expenseConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/expense_db?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
	)

	expensePool, err := pgxpool.New(ctx, expenseConnString)
	if err != nil {
		fmt.Printf("  ERROR: Cannot connect to expense_db: %v\n", err)
	} else {
		defer expensePool.Close()

		// Check if expenses table exists
		var tableExists bool
		err = expensePool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'expenses')").Scan(&tableExists)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else if tableExists {
			fmt.Println("  OK: expenses table exists")

			// Get table columns
			rows, err := expensePool.Query(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'expenses' ORDER BY ordinal_position")
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
			rows, err = expensePool.Query(ctx, "SELECT indexname FROM pg_indexes WHERE tablename = 'expenses' ORDER BY indexname")
			if err == nil {
				fmt.Println("  Indexes:")
				for rows.Next() {
					var indexName string
					rows.Scan(&indexName)
					fmt.Printf("    - %s\n", indexName)
				}
				rows.Close()
			}

			// Count expenses
			var expenseCount int
			expensePool.QueryRow(ctx, "SELECT COUNT(*) FROM expenses WHERE deleted_at IS NULL").Scan(&expenseCount)
			fmt.Printf("  Active expenses: %d\n", expenseCount)
		} else {
			fmt.Println("  WARNING: expenses table does not exist")
		}
	}

	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("Database check completed!")
	fmt.Println("========================================")
}
