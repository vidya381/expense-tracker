package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/vidya381/expense-tracker-backend/models"
)

// AddBudget creates a new budget for a user
func AddBudget(db *sql.DB, budget models.Budget) error {
	period := budget.Period
	if period != "monthly" && period != "yearly" {
		return fmt.Errorf("period must be monthly or yearly")
	}

	// Validate alert threshold
	if budget.AlertThreshold < 0 || budget.AlertThreshold > 100 {
		return fmt.Errorf("alert threshold must be between 0 and 100")
	}

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO budgets (user_id, category_id, amount, period, alert_threshold)
		 VALUES ($1, $2, $3, $4, $5)`,
		budget.UserID, budget.CategoryID, budget.Amount, period, budget.AlertThreshold)
	return err
}

// ListBudgets retrieves all budgets for a user with current spending
func ListBudgets(db *sql.DB, userID int) ([]models.Budget, error) {
	rows, err := db.QueryContext(context.Background(),
		`SELECT b.id, b.user_id, b.category_id, b.amount, b.period, b.alert_threshold, b.created_at,
		        COALESCE(c.name, 'Overall') as category_name
		 FROM budgets b
		 LEFT JOIN categories c ON b.category_id = c.id
		 WHERE b.user_id = $1
		 ORDER BY b.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []models.Budget
	for rows.Next() {
		var b models.Budget
		var createdAt time.Time
		err := rows.Scan(&b.ID, &b.UserID, &b.CategoryID, &b.Amount, &b.Period,
			&b.AlertThreshold, &createdAt, &b.CategoryName)
		if err != nil {
			return nil, err
		}
		b.CreatedAt = createdAt.Format("2006-01-02")

		// Calculate current spending for this budget period
		spending, err := calculateCurrentSpending(db, userID, b.CategoryID, b.Period)
		if err != nil {
			return nil, err
		}
		b.CurrentSpending = spending

		budgets = append(budgets, b)
	}
	return budgets, nil
}

// calculateCurrentSpending calculates spending for the current period
func calculateCurrentSpending(db *sql.DB, userID int, categoryID int, period string) (float64, error) {
	now := time.Now()
	var startDate, endDate time.Time

	if period == "monthly" {
		// Current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)
	} else {
		// Current year
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(1, 0, 0).Add(-time.Second)
	}

	var query string
	var args []interface{}

	if categoryID == 0 {
		// Overall budget - sum all expenses
		query = `SELECT COALESCE(SUM(t.amount), 0)
		         FROM transactions t
		         JOIN categories c ON t.category_id = c.id
		         WHERE t.user_id = $1 AND c.type = 'expense'
		         AND t.date >= $2 AND t.date <= $3`
		args = []interface{}{userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
	} else {
		// Category-specific budget
		query = `SELECT COALESCE(SUM(amount), 0)
		         FROM transactions
		         WHERE user_id = $1 AND category_id = $2
		         AND date >= $3 AND date <= $4`
		args = []interface{}{userID, categoryID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")}
	}

	var spending float64
	err := db.QueryRowContext(context.Background(), query, args...).Scan(&spending)
	return spending, err
}

// UpdateBudget updates an existing budget
func UpdateBudget(db *sql.DB, userID, budgetID int, amount float64, alertThreshold int) error {
	if alertThreshold < 0 || alertThreshold > 100 {
		return fmt.Errorf("alert threshold must be between 0 and 100")
	}

	result, err := db.ExecContext(context.Background(),
		`UPDATE budgets
		 SET amount = $1, alert_threshold = $2
		 WHERE id = $3 AND user_id = $4`,
		amount, alertThreshold, budgetID, userID)
	if err != nil {
		return err
	}

	// Check if any rows were actually updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("budget not found or unauthorized")
	}

	return nil
}

// DeleteBudget removes a budget
func DeleteBudget(db *sql.DB, budgetID, userID int) error {
	result, err := db.ExecContext(context.Background(),
		`DELETE FROM budgets WHERE id = $1 AND user_id = $2`,
		budgetID, userID)
	if err != nil {
		return err
	}

	// Check if any rows were actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("budget not found or unauthorized")
	}

	return nil
}

// GetBudgetAlerts returns budgets that have exceeded their alert threshold
func GetBudgetAlerts(db *sql.DB, userID int) ([]models.Budget, error) {
	budgets, err := ListBudgets(db, userID)
	if err != nil {
		return nil, err
	}

	var alerts []models.Budget
	for _, b := range budgets {
		// Skip if budget amount is zero to prevent division by zero
		if b.Amount == 0 {
			continue
		}
		percentage := (b.CurrentSpending / b.Amount) * 100
		if percentage >= float64(b.AlertThreshold) {
			alerts = append(alerts, b)
		}
	}
	return alerts, nil
}
