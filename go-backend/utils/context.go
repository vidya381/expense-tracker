package utils

import (
	"context"
	"time"
)

// DBContext creates a context with database operation timeout from a parent context.
// If parent is nil, uses context.Background().
// Returns the context and a cancel function that should be called when done.
func DBContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, 10*time.Second)
}
