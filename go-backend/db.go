package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/vidya381/myspendo-backend/utils"
)

func connectDB() (*pgx.Conn, error) {
	err := godotenv.Load()
	if err != nil {
		slog.Warn(".env file not found, using system environment variables")
	}

	// Validate database configuration
	if err := utils.ValidateDBConfig(); err != nil {
		return nil, err
	}

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	return conn, nil
}
