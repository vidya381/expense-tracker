package models

type Transaction struct {
	ID           int     `json:"id"`
	UserID       int     `json:"user_id" validate:"required,gt=0"`
	CategoryID   int     `json:"category_id" validate:"required,gt=0"`
	CategoryName string  `json:"category"`
	CategoryType string  `json:"category_type"` // "income" or "expense"
	Amount       float64 `json:"amount" validate:"required,gt=0"`
	Description  string  `json:"description" validate:"max=500"`
	Date         string  `json:"date" validate:"required"`
	CreatedAt    string  `json:"created_at"`
}
