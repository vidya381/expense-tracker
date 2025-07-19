package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vidya381/expense-tracker-backend/models"
)

// Validate the recurrence field and insert if valid
func AddRecurringTransaction(db *sql.DB, rt models.RecurringTransaction) error {
	rec := strings.ToLower(rt.Recurrence)
	if rec != "daily" && rec != "weekly" && rec != "monthly" && rec != "yearly" {
		return fmt.Errorf("recurrence must be daily, weekly, monthly, or yearly")
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO recurring_transactions
		(user_id, category_id, amount, description, start_date, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		rt.UserID, rt.CategoryID, rt.Amount, rt.Description, rt.StartDate, rec)
	return err
}
