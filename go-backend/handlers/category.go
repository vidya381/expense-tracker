package handlers

import (
	"context"
	"database/sql"

	"github.com/vidya381/expense-tracker-backend/models"
)

// AddCategory inserts a new category for the user
func AddCategory(db *sql.DB, userID int, name, ctype string) error {
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO categories (user_id, name, type) VALUES ($1, $2, $3)",
		userID, name, ctype)
	return err
}

// ListCategories fetches all categories for the user
func ListCategories(db *sql.DB, userID int) ([]models.Category, error) {
	rows, err := db.QueryContext(context.Background(),
		"SELECT id, user_id, name, type, created_at FROM categories WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		if err := rows.Scan(&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, nil
}
