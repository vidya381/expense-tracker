package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/vidya381/expense-tracker-backend/utils"
)

func main() {
	// Initialize structured logger
	utils.InitLogger()

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found")
	}

	// Validate required database environment variables
	if err := utils.ValidateDBConfig(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		os.Exit(1)
	}

	// Connect to database
	db, err := sql.Open("pgx", getDBConnURL())
	if err != nil {
		slog.Error("Failed to open database connection", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	slog.Info("Connected to PostgreSQL successfully")

	// Read and execute migration file
	migrationFile := "migrations/001_create_budgets_table.sql"
	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		slog.Error("Failed to read migration file", "file", migrationFile, "error", err)
		os.Exit(1)
	}

	sqlStatement := string(sqlBytes)
	_, err = db.Exec(sqlStatement)
	if err != nil {
		slog.Error("Failed to execute migration", "error", err)
		os.Exit(1)
	}

	slog.Info("Migration applied successfully")
	slog.Info("Budgets table created")
}

func getDBConnURL() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}
