package handlers

import (
	"database/sql"
	"fmt"

	"github.com/vidya381/expense-tracker-backend/models"
	"github.com/vidya381/expense-tracker-backend/utils"
)

// AddCategory inserts a new category for the user
func AddCategory(db *sql.DB, userID int, name, ctype string) (int, error) {
	ctx, cancel := utils.DBContext()
	defer cancel()

	// Check if category already exists for this user/type
	var exists int
	err := db.QueryRowContext(
		ctx,
		"SELECT 1 FROM categories WHERE user_id=$1 AND name=$2 AND type=$3",
		userID, name, ctype).Scan(&exists)
	if err == nil {
		return 0, fmt.Errorf("category '%s' (type: %s) already exists for this user", name, ctype)
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	// Insert category and get the returning ID
	var categoryID int
	err = db.QueryRowContext(
		ctx,
		"INSERT INTO categories (user_id, name, type) VALUES ($1, $2, $3) RETURNING id",
		userID, name, ctype).Scan(&categoryID)
	if err != nil {
		return 0, err
	}

	return categoryID, nil
}

// ListCategories fetches all categories for the user
func ListCategories(db *sql.DB, userID int) ([]models.Category, error) {
	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx,
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
