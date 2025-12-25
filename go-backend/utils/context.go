package utils

import (
	"context"
	"time"
)

// DBContext creates a new context with a standard database operation timeout
// Returns the context and a cancel function that should be called when done
func DBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
