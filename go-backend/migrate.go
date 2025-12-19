package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Connect to database
	db, err := sql.Open("pgx", getDBConnURL())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		panic("Failed to ping DB: " + err.Error())
	}

	fmt.Println("Connected to PostgreSQL successfully!")

	// Read and execute migration file
	migrationFile := "migrations/001_create_budgets_table.sql"
	sqlBytes, err := os.ReadFile(migrationFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to read migration file: %v", err))
	}

	sqlStatement := string(sqlBytes)
	_, err = db.Exec(sqlStatement)
	if err != nil {
		panic(fmt.Sprintf("Failed to execute migration: %v", err))
	}

	fmt.Println("✓ Migration applied successfully!")
	fmt.Println("✓ Budgets table created")
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
