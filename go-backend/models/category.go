package models

type Category struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // "income" or "expense"
	CreatedAt string `json:"created_at"`
}
