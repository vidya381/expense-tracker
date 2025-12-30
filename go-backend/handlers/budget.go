package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/vidya381/expense-tracker-backend/constants"
	"github.com/vidya381/expense-tracker-backend/models"
	"github.com/vidya381/expense-tracker-backend/utils"
)

// AddBudget creates a new budget for a user
func AddBudget(ctx context.Context, db *sql.DB, budget models.Budget) error {
	period := budget.Period
	if period != "monthly" && period != "yearly" {
		return fmt.Errorf("period must be monthly or yearly")
	}

	// Validate alert threshold
	if budget.AlertThreshold < constants.MinAlertThreshold || budget.AlertThreshold > constants.MaxAlertThreshold {
		return fmt.Errorf("alert threshold must be between %d and %d", constants.MinAlertThreshold, constants.MaxAlertThreshold)
	}

	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	_, err := db.ExecContext(ctx,
		`INSERT INTO budgets (user_id, category_id, amount, period, alert_threshold)
		 VALUES ($1, $2, $3, $4, $5)`,
		budget.UserID, budget.CategoryID, budget.Amount, period, budget.AlertThreshold)
	if err != nil {
		// Check for duplicate key constraint violation (PostgreSQL error code 23505)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "23505") {
			return fmt.Errorf("budget already exists for this category and period")
		}
		return fmt.Errorf("failed to insert budget: %w", err)
	}
	return nil
}

// ListBudgets retrieves all budgets for a user with current spending
func ListBudgets(ctx context.Context, db *sql.DB, userID int) ([]models.Budget, error) {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	// Use UTC for all date calculations to avoid timezone issues
	now := time.Now().UTC()
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	currentMonthEnd := currentMonthStart.AddDate(0, 1, 0).Add(-time.Second)
	currentYearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	currentYearEnd := currentYearStart.AddDate(1, 0, 0).Add(-time.Second)

	// Single query with lateral join to calculate spending for all budgets at once
	query := `
		SELECT
			b.id,
			b.user_id,
			b.category_id,
			b.amount,
			b.period,
			b.alert_threshold,
			b.created_at,
			COALESCE(c.name, 'Overall') as category_name,
			COALESCE(spending.total, 0) as current_spending
		FROM budgets b
		LEFT JOIN categories c ON b.category_id = c.id
		LEFT JOIN LATERAL (
			SELECT COALESCE(SUM(t.amount), 0) as total
			FROM transactions t
			JOIN categories cat ON t.category_id = cat.id
			WHERE t.user_id = b.user_id
				AND (
					-- Category-specific budget
					(b.category_id > 0 AND t.category_id = b.category_id) OR
					-- Overall budget (all expenses)
					(b.category_id = 0 AND cat.type = 'expense')
				)
				AND (
					-- Monthly period
					(b.period = 'monthly' AND t.date >= $2 AND t.date <= $3) OR
					-- Yearly period
					(b.period = 'yearly' AND t.date >= $4 AND t.date <= $5)
				)
		) spending ON true
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC`

	rows, err := db.QueryContext(ctx, query, userID,
		currentMonthStart.Format("2006-01-02"), currentMonthEnd.Format("2006-01-02"),
		currentYearStart.Format("2006-01-02"), currentYearEnd.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate for typical number of budgets
	budgets := make([]models.Budget, 0, constants.TypicalBudgetCount)
	for rows.Next() {
		var b models.Budget
		var createdAt time.Time
		err := rows.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Period,
			&b.AlertThreshold, &createdAt, &b.CategoryName, &b.CurrentSpending)
		if err != nil {
			return nil, err
		}
		b.CreatedAt = createdAt.Format("2006-01-02")
		budgets = append(budgets, b)
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return budgets, nil
}

// UpdateBudget modifies an existing budget's amount and alert threshold.
// Verifies that the budget belongs to the user before updating.
// Returns an error if the budget doesn't exist or belongs to another user.
func UpdateBudget(ctx context.Context, db *sql.DB, userID, budgetID int, amount float64, alertThreshold int) error {
	if alertThreshold < constants.MinAlertThreshold || alertThreshold > constants.MaxAlertThreshold {
		return fmt.Errorf("alert threshold must be between %d and %d", constants.MinAlertThreshold, constants.MaxAlertThreshold)
	}

	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	result, err := db.ExecContext(ctx,
		`UPDATE budgets
		 SET amount = $1, alert_threshold = $2
		 WHERE id = $3 AND user_id = $4`,
		amount, alertThreshold, budgetID, userID)
	if err != nil {
		return fmt.Errorf("failed to update budget: %w", err)
	}

	// Check if any rows were actually updated
	return utils.CheckRowsAffected(result, "budget")
}

// DeleteBudget removes a budget from the database.
// Returns an error if the budget doesn't exist or belongs to another user.
func DeleteBudget(ctx context.Context, db *sql.DB, budgetID, userID int) error {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	result, err := db.ExecContext(ctx,
		`DELETE FROM budgets WHERE id = $1 AND user_id = $2`,
		budgetID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	// Check if any rows were actually deleted
	return utils.CheckRowsAffected(result, "budget")
}

// GetBudgetAlerts retrieves all budgets that have exceeded their alert threshold percentage.
// Returns only budgets where current spending is at or above the configured alert level.
func GetBudgetAlerts(ctx context.Context, db *sql.DB, userID int) ([]models.Budget, error) {
	budgets, err := ListBudgets(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	var alerts []models.Budget
	for _, b := range budgets {
		// Skip if budget amount is zero to prevent division by zero
		if b.Amount == 0 {
			continue
		}
		percentage := (b.CurrentSpending / b.Amount) * float64(constants.MaxAlertThreshold)
		if percentage >= float64(b.AlertThreshold) {
			alerts = append(alerts, b)
		}
	}
	return alerts, nil
}
