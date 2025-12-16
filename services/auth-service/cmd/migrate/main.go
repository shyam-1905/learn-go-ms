//go:build migrate

package main

import (
	"context"
	"expense-tracker/auth-service/internal/config"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	runMigration()
}

func runMigration() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required database config
	if cfg.DBHost == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" {
		log.Fatal("Database configuration is incomplete. Please set DB_HOST, DB_USER, DB_PASSWORD, and DB_NAME environment variables.")
	}

	// First, connect to default 'postgres' database to create our database if needed
	defaultConnString := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
	)

	log.Println("Connecting to PostgreSQL server...")
	log.Printf("Host: %s:%d", cfg.DBHost, cfg.DBPort)
	log.Printf("User: %s", cfg.DBUser)

	// Connect to default postgres database
	defaultPool, err := pgxpool.New(context.Background(), defaultConnString)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL server: %v", err)
	}
	defer defaultPool.Close()

	// Check if database exists and create if needed
	ctx := context.Background()
	var dbExists bool
	err = defaultPool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.DBName).Scan(&dbExists)

	if err != nil {
		log.Fatalf("Failed to check if database exists: %v", err)
	}

	if !dbExists {
		log.Printf("Database '%s' does not exist. Creating it...", cfg.DBName)
		// Note: We can't use parameterized query for CREATE DATABASE
		// So we need to be careful with the database name (it's from config, should be safe)
		_, err = defaultPool.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		log.Printf("✅ Database '%s' created successfully!", cfg.DBName)
	} else {
		log.Printf("✅ Database '%s' already exists", cfg.DBName)
	}

	// Close default connection
	defaultPool.Close()

	// Now connect to our target database
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=require",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	log.Printf("Connecting to database '%s'...", cfg.DBName)

	// Create connection pool to target database
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connection established!")

	// Get migration file path
	migrationFile := "migrations/001_create_users_table.sql"
	if len(os.Args) > 1 {
		migrationFile = os.Args[1]
	}

	// Resolve absolute path
	migrationPath, err := filepath.Abs(migrationFile)
	if err != nil {
		log.Fatalf("Failed to resolve migration file path: %v", err)
	}

	log.Printf("Reading migration file: %s", migrationPath)

	// Read migration file
	sqlBytes, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	sql := string(sqlBytes)

	// Remove comments and empty lines for cleaner execution
	// Split by semicolons and execute each statement
	statements := strings.Split(sql, ";")

	executed := 0

	for i, stmt := range statements {
		// Remove single-line comments (lines starting with --)
		lines := strings.Split(stmt, "\n")
		var cleanLines []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Skip comment-only lines, but keep lines with code before comments
			if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				// Remove inline comments (-- at end of line)
				if idx := strings.Index(line, "--"); idx != -1 {
					line = strings.TrimSpace(line[:idx])
				}
				if line != "" {
					cleanLines = append(cleanLines, line)
				}
			}
		}
		stmt = strings.Join(cleanLines, "\n")
		stmt = strings.TrimSpace(stmt)

		// Skip empty statements
		if stmt == "" {
			continue
		}

		// Execute statement
		log.Printf("Executing statement %d...", i+1)
		_, err := pool.Exec(ctx, stmt)
		if err != nil {
			// Check if it's a "already exists" error (which is okay)
			if strings.Contains(err.Error(), "already exists") {
				log.Printf("⚠️  Statement %d: %v (skipping - already exists)", i+1, err)
				continue
			}
			log.Fatalf("❌ Failed to execute statement %d: %v\nStatement: %s", i+1, err, stmt)
		}
		executed++
		log.Printf("✅ Statement %d executed successfully", i+1)
	}

	log.Println("")
	log.Printf("✅ Migration completed! Executed %d statements.", executed)
	log.Println("")
	log.Println("You can now start the auth-service with: go run cmd/main.go")
}
