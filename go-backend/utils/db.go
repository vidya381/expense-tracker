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
