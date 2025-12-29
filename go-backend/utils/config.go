package utils

import (
	"fmt"
	"os"
	"strings"
)

// ValidateConfig checks that all required environment variables are set
// Returns an error with details about missing variables
// Accepts either DATABASE_URL or individual DB variables (DB_HOST, DB_PORT, etc.)
func ValidateConfig() error {
	// Check JWT_SECRET (always required)
	if os.Getenv("JWT_SECRET") == "" {
		return fmt.Errorf("missing required environment variable: JWT_SECRET")
	}

	// Check database configuration (either DATABASE_URL or individual variables)
	if err := ValidateDBConfig(); err != nil {
		return err
	}

	return nil
}

// ValidateDBConfig checks that all required database environment variables are set
// Used by tools that only need database access (like migrations)
// Accepts either DATABASE_URL or individual DB variables (DB_HOST, DB_PORT, etc.)
func ValidateDBConfig() error {
	// If DATABASE_URL is provided, we're good
	if os.Getenv("DATABASE_URL") != "" {
		return nil
	}

	// Otherwise, check for individual variables
	required := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
	}

	var missing []string
	for _, key := range required {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required database environment variables: %s (or set DATABASE_URL)", strings.Join(missing, ", "))
	}

	return nil
}
