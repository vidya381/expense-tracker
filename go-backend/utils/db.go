package utils

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

// IsConnectionError checks if an error is related to database connectivity issues
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common connection-related errors
	errMsg := err.Error()
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no such host",
		"i/o timeout",
		"eof",
		"connection timed out",
		"network is unreachable",
		"too many connections",
		"connection pool exhausted",
	}

	errMsgLower := strings.ToLower(errMsg)
	for _, connErr := range connectionErrors {
		if strings.Contains(errMsgLower, connErr) {
			return true
		}
	}

	// Check for sql.ErrConnDone
	if errors.Is(err, sql.ErrConnDone) {
		return true
	}

	return false
}

// RetryableDBOperation executes a database operation with retry logic for connection failures
// maxRetries: maximum number of retry attempts (typically 3)
// operation: the database operation to execute
func RetryableDBOperation(maxRetries int, operation func() error) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}

		// Only retry on connection errors
		if !IsConnectionError(err) {
			return err
		}

		// Don't sleep on the last attempt
		if attempt < maxRetries {
			// Exponential backoff: 100ms, 200ms, 400ms
			backoff := time.Duration(100*(1<<uint(attempt))) * time.Millisecond
			time.Sleep(backoff)
		}
	}

	// All retries exhausted, return connection error with helpful message
	return errors.New("database connection unavailable. Please try again later")
}

// VerifyCategoryOwnership checks if a category belongs to the specified user.
// Returns nil if the category is valid and belongs to the user, or if categoryID is 0 (no category).
// Returns an error if the category doesn't exist or belongs to another user.
func VerifyCategoryOwnership(db *sql.DB, userID, categoryID int) error {
	if categoryID == 0 {
		return nil // Allow no category
	}
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE id = $1 AND user_id = $2",
		categoryID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("category not found or unauthorized")
	}
	return nil
}

// CheckRowsAffected verifies that a database operation affected at least one row.
// Returns an error if no rows were affected (typically meaning the record doesn't exist or user is unauthorized).
func CheckRowsAffected(result sql.Result, resourceName string) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New(resourceName + " not found or unauthorized")
	}
	return nil
}
