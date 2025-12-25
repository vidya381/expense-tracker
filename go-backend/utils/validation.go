package utils

import (
	"fmt"
	"time"
)

// ValidateDate checks if a date string is in YYYY-MM-DD format and is a valid date
func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return fmt.Errorf("date is required")
	}

	// Parse the date in YYYY-MM-DD format
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)")
	}

	// Check if the date is not too far in the past (more than 10 years)
	tenYearsAgo := time.Now().AddDate(-10, 0, 0)
	if parsedDate.Before(tenYearsAgo) {
		return fmt.Errorf("date cannot be more than 10 years in the past")
	}

	// Check if the date is not in the future (more than 1 day)
	tomorrow := time.Now().AddDate(0, 0, 1)
	if parsedDate.After(tomorrow) {
		return fmt.Errorf("date cannot be in the future")
	}

	return nil
}

// ValidateTransactionDate validates date for transactions with more lenient rules
func ValidateTransactionDate(dateStr string) error {
	if dateStr == "" {
		return fmt.Errorf("date is required")
	}

	// Parse the date in YYYY-MM-DD format
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid date format. Expected YYYY-MM-DD (e.g., 2025-12-25)")
	}

	return nil
}

// ValidateRecurringDate validates date for recurring transactions
func ValidateRecurringDate(dateStr string) error {
	if dateStr == "" {
		return fmt.Errorf("start date is required")
	}

	// Parse the date in YYYY-MM-DD format
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("invalid start date format. Expected YYYY-MM-DD (e.g., 2025-12-25)")
	}

	// Start date should not be too far in the future (more than 1 year)
	oneYearFromNow := time.Now().AddDate(1, 0, 0)
	if parsedDate.After(oneYearFromNow) {
		return fmt.Errorf("start date cannot be more than 1 year in the future")
	}

	// Start date should not be too far in the past (more than 5 years)
	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	if parsedDate.Before(fiveYearsAgo) {
		return fmt.Errorf("start date cannot be more than 5 years in the past")
	}

	return nil
}

// ValidateAmount checks if an amount is valid (positive and reasonable)
func ValidateAmount(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Check for unreasonably large amounts (> 1 billion)
	if amount > 1000000000 {
		return fmt.Errorf("amount is too large. Maximum allowed is 1,000,000,000")
	}

	return nil
}

// ValidatePaginationParams validates limit and offset for pagination
func ValidatePaginationParams(limit, offset int) error {
	if limit < 1 {
		return fmt.Errorf("limit must be at least 1")
	}

	if limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000 records")
	}

	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	return nil
}
