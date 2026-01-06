package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/vidya381/myspendo-backend/models"
	"github.com/vidya381/myspendo-backend/utils"
)

// AddRecurringTransaction creates a new recurring transaction that automatically generates transactions.
// Validates that the recurrence is 'daily', 'weekly', 'monthly', or 'yearly' and that the category belongs to the user.
// Recurring transactions are processed by a background job to create actual transactions.
func AddRecurringTransaction(ctx context.Context, db *sql.DB, rt models.RecurringTransaction) error {
	rec := strings.ToLower(rt.Recurrence)
	if rec != "daily" && rec != "weekly" && rec != "monthly" && rec != "yearly" {
		return fmt.Errorf("recurrence must be daily, weekly, monthly, or yearly")
	}

	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	// Verify category ownership before creating recurring transaction
	if err := utils.VerifyCategoryOwnership(db, rt.UserID, rt.CategoryID); err != nil {
		return err
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO recurring_transactions
		(user_id, category_id, amount, description, start_date, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		rt.UserID, rt.CategoryID, rt.Amount, rt.Description, rt.StartDate, rec)
	if err != nil {
		return fmt.Errorf("failed to insert recurring transaction: %w", err)
	}
	return nil
}

// ListRecurringTransactions retrieves all recurring transactions for the specified user.
// Includes information about when each recurring transaction was last processed.
func ListRecurringTransactions(ctx context.Context, db *sql.DB, userID int) ([]models.RecurringTransaction, error) {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, category_id, amount, description, start_date, recurrence, last_occurrence, created_at
		 FROM recurring_transactions
		 WHERE user_id = $1
		 ORDER BY start_date DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate for typical number of recurring transactions (5-20)
	list := make([]models.RecurringTransaction, 0, 10)
	for rows.Next() {
		var rt models.RecurringTransaction
		var lastOccurrence sql.NullTime
		var createdAt time.Time
		err := rows.Scan(&rt.ID, &rt.UserID, &rt.CategoryID, &rt.Amount, &rt.Description, &rt.StartDate, &rt.Recurrence, &lastOccurrence, &createdAt)
		if err != nil {
			return nil, err
		}
		if lastOccurrence.Valid {
			rt.LastOccurrence = &lastOccurrence.Time
		} else {
			rt.LastOccurrence = nil
		}
		rt.CreatedAt = createdAt.Format("2006-01-02")
		list = append(list, rt)
	}
	return list, nil
}

// EditRecurringTransaction updates an existing recurring transaction's amount, description, start date, and recurrence.
// Verifies that the recurring transaction belongs to the user before updating.
// Returns an error if the transaction doesn't exist or belongs to another user.
func EditRecurringTransaction(ctx context.Context, db *sql.DB, userID, id int, amount float64, description, startDate, recurrence string) error {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	// Only allow update if user owns it
	result, err := db.ExecContext(ctx,
		`UPDATE recurring_transactions
		 SET amount = $1, description = $2, start_date = $3, recurrence = $4
		 WHERE id = $5 AND user_id = $6`,
		amount, description, startDate, recurrence, id, userID)
	if err != nil {
		return fmt.Errorf("failed to update recurring transaction: %w", err)
	}

	// Check if any rows were actually updated
	return utils.CheckRowsAffected(result, "recurring transaction")
}

// DeleteRecurringTransaction removes a recurring transaction from the database.
// Returns an error if the transaction doesn't exist or belongs to another user.
// Note: This does not delete the transactions that were already created from this recurring rule.
func DeleteRecurringTransaction(ctx context.Context, db *sql.DB, id, userID int) error {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	result, err := db.ExecContext(ctx,
		"DELETE FROM recurring_transactions WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete recurring transaction: %w", err)
	}

	// Check if any rows were actually deleted
	return utils.CheckRowsAffected(result, "recurring transaction")
}
