package models

type Category struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id" validate:"required,gt=0"`
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Type      string `json:"type" validate:"required,oneof=income expense"`
	CreatedAt string `json:"created_at"`
}
