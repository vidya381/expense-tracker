package models

import "time"

type RecurringTransaction struct {
	ID             int
	UserID         int
	CategoryID     int
	Amount         float64
	Description    string
	StartDate      string
	Recurrence     string // daily, weekly, monthly, yearly
	LastOccurrence *time.Time
	CreatedAt      string
}
