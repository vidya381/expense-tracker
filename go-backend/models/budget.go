package models

type Budget struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id" validate:"required,gt=0"`
	CategoryID      int     `json:"category_id" validate:"gte=0"` // 0 means overall budget
	CategoryName    string  `json:"category_name,omitempty"`
	Amount          float64 `json:"amount" validate:"required,gt=0"`
	Period          string  `json:"period" validate:"required,oneof=monthly yearly"`
	AlertThreshold  int     `json:"alert_threshold" validate:"required,gte=0,lte=100"` // percentage (e.g., 80 means alert at 80%)
	CurrentSpending float64 `json:"current_spending"`                                   // calculated, not stored
	CreatedAt       string  `json:"created_at"`
}
