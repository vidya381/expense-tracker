package utils

import (
	"testing"
	"time"
)

func TestValidateDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date - today",
			date:    time.Now().UTC().Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "valid date - yesterday",
			date:    time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "valid date - one month ago",
			date:    time.Now().UTC().AddDate(0, -1, 0).Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "invalid date - empty string",
			date:    "",
			wantErr: true,
		},
		{
			name:    "invalid date - wrong format",
			date:    "12/25/2025",
			wantErr: true,
		},
		{
			name:    "invalid date - wrong format 2",
			date:    "25-12-2025",
			wantErr: true,
		},
		{
			name:    "invalid date - future date",
			date:    time.Now().UTC().AddDate(0, 0, 2).Format("2006-01-02"),
			wantErr: true,
		},
		{
			name:    "invalid date - too far in past",
			date:    time.Now().UTC().AddDate(-11, 0, 0).Format("2006-01-02"),
			wantErr: true,
		},
		{
			name:    "invalid date - invalid day",
			date:    "2025-02-30",
			wantErr: true,
		},
		{
			name:    "invalid date - invalid month",
			date:    "2025-13-01",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTransactionDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date - today",
			date:    "2025-12-25",
			wantErr: false,
		},
		{
			name:    "valid date - past",
			date:    "2020-01-01",
			wantErr: false,
		},
		{
			name:    "valid date - future",
			date:    "2030-12-31",
			wantErr: false,
		},
		{
			name:    "invalid date - empty string",
			date:    "",
			wantErr: true,
		},
		{
			name:    "invalid date - wrong format",
			date:    "12/25/2025",
			wantErr: true,
		},
		{
			name:    "invalid date - invalid day",
			date:    "2025-02-30",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTransactionDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTransactionDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRecurringDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{
			name:    "valid date - today",
			date:    time.Now().UTC().Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "valid date - 6 months from now",
			date:    time.Now().UTC().AddDate(0, 6, 0).Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "valid date - 1 month ago",
			date:    time.Now().UTC().AddDate(0, -1, 0).Format("2006-01-02"),
			wantErr: false,
		},
		{
			name:    "invalid date - empty string",
			date:    "",
			wantErr: true,
		},
		{
			name:    "invalid date - too far in future",
			date:    time.Now().UTC().AddDate(2, 0, 0).Format("2006-01-02"),
			wantErr: true,
		},
		{
			name:    "invalid date - too far in past",
			date:    time.Now().UTC().AddDate(-6, 0, 0).Format("2006-01-02"),
			wantErr: true,
		},
		{
			name:    "invalid date - wrong format",
			date:    "25-12-2025",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecurringDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecurringDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAmount(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		wantErr bool
	}{
		{
			name:    "valid amount - small",
			amount:  0.01,
			wantErr: false,
		},
		{
			name:    "valid amount - medium",
			amount:  100.50,
			wantErr: false,
		},
		{
			name:    "valid amount - large",
			amount:  999999999,
			wantErr: false,
		},
		{
			name:    "invalid amount - zero",
			amount:  0,
			wantErr: true,
		},
		{
			name:    "invalid amount - negative",
			amount:  -10.50,
			wantErr: true,
		},
		{
			name:    "invalid amount - too large",
			amount:  1000000001,
			wantErr: true,
		},
		{
			name:    "invalid amount - extremely large",
			amount:  9999999999999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAmount(tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAmount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePaginationParams(t *testing.T) {
	tests := []struct {
		name    string
		limit   int
		offset  int
		wantErr bool
	}{
		{
			name:    "valid pagination - default",
			limit:   20,
			offset:  0,
			wantErr: false,
		},
		{
			name:    "valid pagination - with offset",
			limit:   50,
			offset:  100,
			wantErr: false,
		},
		{
			name:    "valid pagination - max limit",
			limit:   1000,
			offset:  0,
			wantErr: false,
		},
		{
			name:    "invalid pagination - limit too small",
			limit:   0,
			offset:  0,
			wantErr: true,
		},
		{
			name:    "invalid pagination - limit negative",
			limit:   -1,
			offset:  0,
			wantErr: true,
		},
		{
			name:    "invalid pagination - limit too large",
			limit:   1001,
			offset:  0,
			wantErr: true,
		},
		{
			name:    "invalid pagination - offset negative",
			limit:   20,
			offset:  -1,
			wantErr: true,
		},
		{
			name:    "invalid pagination - both invalid",
			limit:   -1,
			offset:  -1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePaginationParams(tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePaginationParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
