package utils

import (
	"database/sql"
	"errors"
	"testing"
)

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		want  bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "connection refused",
			err:  errors.New("connection refused"),
			want: true,
		},
		{
			name: "Connection Refused (capitalized)",
			err:  errors.New("dial tcp: Connection Refused"),
			want: true,
		},
		{
			name: "connection reset",
			err:  errors.New("connection reset by peer"),
			want: true,
		},
		{
			name: "broken pipe",
			err:  errors.New("write: broken pipe"),
			want: true,
		},
		{
			name: "no such host",
			err:  errors.New("dial tcp: lookup failed: no such host"),
			want: true,
		},
		{
			name: "i/o timeout",
			err:  errors.New("i/o timeout"),
			want: true,
		},
		{
			name: "EOF error",
			err:  errors.New("EOF"),
			want: true,
		},
		{
			name: "connection timed out",
			err:  errors.New("connection timed out"),
			want: true,
		},
		{
			name: "network is unreachable",
			err:  errors.New("network is unreachable"),
			want: true,
		},
		{
			name: "too many connections",
			err:  errors.New("too many connections"),
			want: true,
		},
		{
			name: "connection pool exhausted",
			err:  errors.New("connection pool exhausted"),
			want: true,
		},
		{
			name: "sql.ErrConnDone",
			err:  sql.ErrConnDone,
			want: true,
		},
		{
			name: "non-connection error - syntax",
			err:  errors.New("syntax error at or near"),
			want: false,
		},
		{
			name: "non-connection error - constraint violation",
			err:  errors.New("duplicate key value violates unique constraint"),
			want: false,
		},
		{
			name: "non-connection error - not found",
			err:  errors.New("record not found"),
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("something went wrong"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("IsConnectionError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetryableDBOperation(t *testing.T) {
	tests := []struct {
		name          string
		operation     func() error
		wantErr       bool
		wantCallCount int
	}{
		{
			name: "successful operation - first try",
			operation: func() error {
				return nil
			},
			wantErr:       false,
			wantCallCount: 1,
		},
		{
			name: "non-retryable error",
			operation: func() error {
				return errors.New("syntax error")
			},
			wantErr:       true,
			wantCallCount: 1,
		},
		{
			name: "retryable error - all retries fail",
			operation: func() error {
				return errors.New("connection refused")
			},
			wantErr:       true,
			wantCallCount: 4, // 1 initial + 3 retries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			operation := func() error {
				callCount++
				return tt.operation()
			}

			err := RetryableDBOperation(3, operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetryableDBOperation() error = %v, wantErr %v", err, tt.wantErr)
			}

			if callCount != tt.wantCallCount {
				t.Errorf("operation called %d times, want %d", callCount, tt.wantCallCount)
			}
		})
	}
}

func TestRetryableDBOperation_SuccessAfterRetry(t *testing.T) {
	// Test that operation succeeds after a few retries
	attemptCount := 0
	operation := func() error {
		attemptCount++
		if attemptCount < 3 {
			return errors.New("connection refused") // Retryable error
		}
		return nil // Success on 3rd attempt
	}

	err := RetryableDBOperation(3, operation)
	if err != nil {
		t.Errorf("RetryableDBOperation() should succeed after retries, got error: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("operation should be called 3 times, called %d times", attemptCount)
	}
}
