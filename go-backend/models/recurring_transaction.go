package models

import "time"

type RecurringTransaction struct {
	ID             int        `json:"id"`
	UserID         int        `json:"user_id" validate:"required,gt=0"`
	CategoryID     int        `json:"category_id" validate:"required,gt=0"`
	Amount         float64    `json:"amount" validate:"required,gt=0"`
	Description    string     `json:"description" validate:"max=500"`
	StartDate      string     `json:"start_date" validate:"required"`
	Recurrence     string     `json:"recurrence" validate:"required,oneof=daily weekly monthly yearly"`
	LastOccurrence *time.Time `json:"last_occurrence,omitempty"`
	CreatedAt      string     `json:"created_at"`
}
