package utils

import (
	"fmt"
	"os"
	"strings"
)

// ValidateConfig checks that all required environment variables are set
// Returns an error with details about missing variables
func ValidateConfig() error {
	required := []string{
		"JWT_SECRET",
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
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidateDBConfig checks that all required database environment variables are set
// Used by tools that only need database access (like migrations)
func ValidateDBConfig() error {
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
		return fmt.Errorf("missing required database environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}
