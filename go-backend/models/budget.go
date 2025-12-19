package models

type Budget struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`
	CategoryID      int     `json:"category_id"` // 0 means overall budget
	CategoryName    string  `json:"category_name,omitempty"`
	Amount          float64 `json:"amount"`
	Period          string  `json:"period"`           // "monthly" or "yearly"
	AlertThreshold  int     `json:"alert_threshold"`  // percentage (e.g., 80 means alert at 80%)
	CurrentSpending float64 `json:"current_spending"` // calculated, not stored
	CreatedAt       string  `json:"created_at"`
}
