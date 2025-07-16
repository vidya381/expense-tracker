package models

type Transaction struct {
	ID          int
	UserID      int
	CategoryID  int
	Amount      float64
	Description string
	Date        string
}
