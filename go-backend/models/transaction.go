package models

type Transaction struct {
	ID           int     `json:"id"`
	UserID       int     `json:"user_id"`
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Date         string  `json:"date"`
	CreatedAt    string  `json:"created_at"`
}
