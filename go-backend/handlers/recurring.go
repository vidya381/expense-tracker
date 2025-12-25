package handlers

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/vidya381/expense-tracker-backend/models"
	"github.com/vidya381/expense-tracker-backend/utils"
)

// Validate the recurrence field and insert if valid
func AddRecurringTransaction(db *sql.DB, rt models.RecurringTransaction) error {
	rec := strings.ToLower(rt.Recurrence)
	if rec != "daily" && rec != "weekly" && rec != "monthly" && rec != "yearly" {
		return fmt.Errorf("recurrence must be daily, weekly, monthly, or yearly")
	}

	ctx, cancel := utils.DBContext()
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO recurring_transactions
		(user_id, category_id, amount, description, start_date, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		rt.UserID, rt.CategoryID, rt.Amount, rt.Description, rt.StartDate, rec)
	return err
}

// Lists all recurring transactions for a user
func ListRecurringTransactions(db *sql.DB, userID int) ([]models.RecurringTransaction, error) {
	ctx, cancel := utils.DBContext()
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

	var list []models.RecurringTransaction
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

// Edits a recurring transaction. Only fields that make sense are updatable.
func EditRecurringTransaction(db *sql.DB, userID, id int, amount float64, description, startDate, recurrence string) error {
	ctx, cancel := utils.DBContext()
	defer cancel()

	// Only allow update if user owns it
	_, err := db.ExecContext(ctx,
		`UPDATE recurring_transactions
		 SET amount = $1, description = $2, start_date = $3, recurrence = $4
		 WHERE id = $5 AND user_id = $6`,
		amount, description, startDate, recurrence, id, userID)
	return err
}

// Delete recurring transaction
func DeleteRecurringTransaction(db *sql.DB, id, userID int) error {
	ctx, cancel := utils.DBContext()
	defer cancel()

	_, err := db.ExecContext(ctx,
		"DELETE FROM recurring_transactions WHERE id = $1 AND user_id = $2", id, userID)
	return err
}
