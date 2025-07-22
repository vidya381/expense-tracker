package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/vidya381/expense-tracker-backend/models"
)

// Create a new transaction for the user
func AddTransaction(db *sql.DB, tx models.Transaction) error {
	query := `INSERT INTO transactions (user_id, category_id, amount, description, date)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(context.Background(), query,
		tx.UserID, tx.CategoryID, tx.Amount, tx.Description, tx.Date)
	return err
}

// Fetch all transactions for a user
func ListTransactions(db *sql.DB, userID int) ([]models.Transaction, error) {
	rows, err := db.QueryContext(context.Background(),
		`SELECT id, user_id, category_id, amount, description, date, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.CategoryID, &tx.Amount,
			&tx.Description, &tx.Date, &tx.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}

// Update an existing transaction
func UpdateTransaction(db *sql.DB, tx models.Transaction) error {
	query := `UPDATE transactions
			  SET amount = $1, description = $2, category_id = $3, date = $4
			  WHERE id = $5 AND user_id = $6`
	_, err := db.ExecContext(context.Background(), query,
		tx.Amount, tx.Description, tx.CategoryID, tx.Date, tx.ID, tx.UserID)
	return err
}

// Delete a transaction
func DeleteTransaction(db *sql.DB, id, userID int) error {
	_, err := db.ExecContext(context.Background(),
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// FilterTransactionsPaginated supports filtering, pagination, and sorting.
func FilterTransactionsPaginated(
	db *sql.DB,
	userID int,
	keyword string,
	categoryID int,
	dateFrom string,
	dateTo string,
	amountMin float64,
	amountMax float64,
	orderBy string,
	limit int,
	offset int,
) ([]models.Transaction, error) {

	base := `SELECT id, user_id, category_id, amount, description, date, created_at
	         FROM transactions WHERE user_id = $1`
	args := []interface{}{userID}
	argpos := 2

	if keyword != "" {
		base += fmt.Sprintf(" AND description ILIKE $%d", argpos)
		args = append(args, "%"+keyword+"%")
		argpos++
	}
	if categoryID > 0 {
		base += fmt.Sprintf(" AND category_id = $%d", argpos)
		args = append(args, categoryID)
		argpos++
	}
	if dateFrom != "" {
		base += fmt.Sprintf(" AND date >= $%d", argpos)
		args = append(args, dateFrom)
		argpos++
	}
	if dateTo != "" {
		base += fmt.Sprintf(" AND date <= $%d", argpos)
		args = append(args, dateTo)
		argpos++
	}
	if amountMin > 0 {
		base += fmt.Sprintf(" AND amount >= $%d", argpos)
		args = append(args, amountMin)
		argpos++
	}
	if amountMax > 0 {
		base += fmt.Sprintf(" AND amount <= $%d", argpos)
		args = append(args, amountMax)
		argpos++
	}

	if orderBy == "" {
		orderBy = "date DESC"
	}
	base += " ORDER BY " + orderBy

	// Add pagination
	base += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argpos, argpos+1)
	args = append(args, limit, offset)

	rows, err := db.QueryContext(context.Background(), base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.CategoryID, &t.Amount, &t.Description, &t.Date, &t.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, t)
	}
	return results, nil
}
