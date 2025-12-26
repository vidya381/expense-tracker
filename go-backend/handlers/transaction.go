package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/vidya381/expense-tracker-backend/constants"
	"github.com/vidya381/expense-tracker-backend/models"
	"github.com/vidya381/expense-tracker-backend/utils"
)

// verifyCategoryOwnership checks if a category belongs to the user
func verifyCategoryOwnership(db *sql.DB, userID, categoryID int) error {
	if categoryID == 0 {
		return nil // Allow no category
	}
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories WHERE id = $1 AND user_id = $2",
		categoryID, userID).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("category not found or unauthorized")
	}
	return nil
}

// AddTransaction creates a new expense or income transaction for the user.
// Verifies that the specified category belongs to the user before creation.
func AddTransaction(db *sql.DB, tx models.Transaction) error {
	// Verify category ownership
	if err := verifyCategoryOwnership(db, tx.UserID, tx.CategoryID); err != nil {
		return err
	}

	ctx, cancel := utils.DBContext()
	defer cancel()

	query := `INSERT INTO transactions (user_id, category_id, amount, description, date)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(ctx, query,
		tx.UserID, tx.CategoryID, tx.Amount, tx.Description, tx.Date)
	return err
}

// ListTransactions retrieves all transactions for the specified user, including category details.
// Returns transactions in descending order by date (newest first).
func ListTransactions(db *sql.DB, userID int) ([]models.Transaction, error) {
	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`
        SELECT
            t.id,
            t.user_id,
            t.category_id,
            c.name AS category_name,
            t.amount,
            t.description,
            t.date,
            t.created_at
        FROM transactions t
        JOIN categories c ON t.category_id = c.id
        WHERE t.user_id = $1
        ORDER BY t.date DESC
        `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate with reasonable capacity
	transactions := make([]models.Transaction, 0, constants.TypicalTransactionCount)
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.CategoryID,
			&tx.CategoryName,
			&tx.Amount,
			&tx.Description,
			&tx.Date,
			&tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// UpdateTransaction modifies an existing transaction's amount, description, category, and date.
// Verifies category ownership and that the transaction belongs to the user.
// Returns an error if the transaction doesn't exist or belongs to another user.
func UpdateTransaction(db *sql.DB, tx models.Transaction) error {
	// Verify category ownership
	if err := verifyCategoryOwnership(db, tx.UserID, tx.CategoryID); err != nil {
		return err
	}

	ctx, cancel := utils.DBContext()
	defer cancel()

	query := `UPDATE transactions
			  SET amount = $1, description = $2, category_id = $3, date = $4
			  WHERE id = $5 AND user_id = $6`
	result, err := db.ExecContext(ctx, query,
		tx.Amount, tx.Description, tx.CategoryID, tx.Date, tx.ID, tx.UserID)
	if err != nil {
		return err
	}

	// Check if any rows were actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("transaction not found or unauthorized")
	}

	return nil
}

// DeleteTransaction removes a transaction from the database.
// Returns an error if the transaction doesn't exist or belongs to another user.
func DeleteTransaction(db *sql.DB, id, userID int) error {
	ctx, cancel := utils.DBContext()
	defer cancel()

	result, err := db.ExecContext(ctx,
		`DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("transaction not found or unauthorized")
	}

	return nil
}

// FilterTransactionsPaginated retrieves transactions with filtering, pagination, and sorting options.
// Supports filtering by keyword (matches description or category name), category ID, date range, and amount range.
// Results can be ordered by 'date' or 'amount' in ascending or descending order.
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

	base := `SELECT
                t.id,
                t.user_id,
                t.category_id,
                c.name AS category_name,
                c.type AS category_type,
                t.amount,
                t.description,
                t.date,
                t.created_at
             FROM transactions t
             JOIN categories c ON t.category_id = c.id
             WHERE t.user_id = $1`
	args := []interface{}{userID}
	argpos := 2

	if keyword != "" {
		base += fmt.Sprintf(" AND t.description ILIKE $%d", argpos)
		args = append(args, "%"+keyword+"%")
		argpos++
	}
	if categoryID > 0 {
		base += fmt.Sprintf(" AND t.category_id = $%d", argpos)
		args = append(args, categoryID)
		argpos++
	}
	if dateFrom != "" {
		base += fmt.Sprintf(" AND t.date >= $%d", argpos)
		args = append(args, dateFrom)
		argpos++
	}
	if dateTo != "" {
		base += fmt.Sprintf(" AND t.date <= $%d", argpos)
		args = append(args, dateTo)
		argpos++
	}
	if amountMin > 0 {
		base += fmt.Sprintf(" AND t.amount >= $%d", argpos)
		args = append(args, amountMin)
		argpos++
	}
	if amountMax > 0 {
		base += fmt.Sprintf(" AND t.amount <= $%d", argpos)
		args = append(args, amountMax)
		argpos++
	}

	// Validate orderBy to prevent SQL injection
	allowedOrders := map[string]bool{
		"t.date ASC":           true,
		"t.date DESC":          true,
		"t.amount ASC":         true,
		"t.amount DESC":        true,
		"t.created_at ASC":     true,
		"t.created_at DESC":    true,
		"t.date ASC, t.created_at DESC":    true,
		"t.date DESC, t.created_at DESC":   true,
		"t.amount ASC, t.created_at DESC":  true,
		"t.amount DESC, t.created_at DESC": true,
	}
	if orderBy == "" {
		orderBy = "t.date DESC"
	}
	if !allowedOrders[orderBy] {
		orderBy = "t.date DESC" // fallback to default if invalid
	}
	base += " ORDER BY " + orderBy

	base += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argpos, argpos+1)
	args = append(args, limit, offset)

	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate slice with capacity hint (limit) for better performance
	results := make([]models.Transaction, 0, limit)
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.CategoryID,
			&t.CategoryName,
			&t.CategoryType,
			&t.Amount,
			&t.Description,
			&t.Date,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, t)
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
