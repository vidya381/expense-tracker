package models

type Category struct {
	ID     int
	UserID int
	Name   string
	Type   string // "income" or "expense"
}
