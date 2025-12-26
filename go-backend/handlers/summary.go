package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vidya381/expense-tracker-backend/utils"
)

// GetTotals calculates the total expenses and income for the specified user across all time.
// Returns two float64 values: total expenses and total income.
func GetTotals(db *sql.DB, userID int) (expenses float64, income float64, err error) {
	ctx, cancel := utils.DBContext()
	defer cancel()

	err = db.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END),0)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1`, userID).Scan(&expenses, &income)
	return
}

// GetMonthlyTotals retrieves monthly aggregated income and expense totals for the user.
// Returns data grouped by month in descending order (most recent first).
func GetMonthlyTotals(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx,
		`SELECT DATE_TRUNC('month', t.date) as month,
				COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END),0) as total_expenses,
				COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END),0) as total_income
		 FROM transactions t
		 JOIN categories c ON t.category_id = c.id
		 WHERE t.user_id = $1
		 GROUP BY month
		 ORDER BY month DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate for ~12 months of data typically
	results := make([]map[string]interface{}, 0, 12)
	for rows.Next() {
		var month time.Time
		var totalExpenses, totalIncome float64
		if err := rows.Scan(&month, &totalExpenses, &totalIncome); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"month":          month.Format("2006-01"),
			"total_expenses": totalExpenses,
			"total_income":   totalIncome,
		})
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetCategoryBreakdown provides a breakdown of spending by category for the specified user.
// Optionally filters by date range using 'from' and 'to' parameters (format: YYYY-MM-DD).
// Returns data grouped by category and type, sorted by type and total amount.
func GetCategoryBreakdown(db *sql.DB, userID int, from, to string) ([]map[string]interface{}, error) {
	base := `SELECT c.name, c.type, COALESCE(SUM(t.amount),0) AS total
	 FROM transactions t
	 JOIN categories c ON t.category_id = c.id
	 WHERE t.user_id = $1`
	params := []interface{}{userID}
	paramIdx := 2
	if from != "" {
		base += fmt.Sprintf(" AND t.date >= $%d", paramIdx)
		params = append(params, from)
		paramIdx++
	}
	if to != "" {
		base += fmt.Sprintf(" AND t.date <= $%d", paramIdx)
		params = append(params, to)
		paramIdx++
	}
	base += " GROUP BY c.name, c.type ORDER BY c.type, total DESC"

	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx, base, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate for typical number of categories (5-20)
	result := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		var name, ctype string
		var total float64
		if err := rows.Scan(&name, &ctype, &total); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"category": name,
			"type":     ctype,
			"total":    total,
		})
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// GetGroupTotals retrieves income and expense totals grouped by time period.
// The granularity parameter must be 'month', 'week', or 'year'.
// Returns data sorted by period in descending order (most recent first).
func GetGroupTotals(db *sql.DB, userID int, granularity string) ([]map[string]interface{}, error) {
	// Strict whitelist validation - explicitly reject invalid values
	allowedGranularities := map[string]bool{
		"month": true,
		"week":  true,
		"year":  true,
	}

	if !allowedGranularities[granularity] {
		return nil, errors.New("invalid granularity: must be 'month', 'week', or 'year'")
	}

	// Safe to use in SQL since we validated against strict whitelist
	sqlQuery := fmt.Sprintf(`
		SELECT DATE_TRUNC('%s', t.date) as period,
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END), 0) as total_income
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
		GROUP BY period ORDER BY period DESC`, granularity)

	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx, sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate based on granularity (52 weeks, 12 months, or ~3 years)
	capacity := 52
	if granularity == "month" {
		capacity = 12
	} else if granularity == "year" {
		capacity = 3
	}
	results := make([]map[string]interface{}, 0, capacity)
	for rows.Next() {
		var period time.Time
		var totalExpenses, totalIncome float64
		if err := rows.Scan(&period, &totalExpenses, &totalIncome); err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"period":         period.Format("2006-01-02"),
			"total_expenses": totalExpenses,
			"total_income":   totalIncome,
		})
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetCategoryMonthSummary provides a category breakdown for a specific month.
// Returns aggregated expenses and income grouped by category for the specified year and month.
func GetCategoryMonthSummary(db *sql.DB, userID int, year, month int) ([]map[string]interface{}, error) {
	query := `
		SELECT c.name, c.type, COALESCE(SUM(t.amount),0)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1 AND EXTRACT(YEAR FROM t.date) = $2 AND EXTRACT(MONTH FROM t.date) = $3
		GROUP BY c.name, c.type
		ORDER BY c.type, SUM(t.amount) DESC`

	ctx, cancel := utils.DBContext()
	defer cancel()

	rows, err := db.QueryContext(ctx, query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Pre-allocate for typical number of categories
	result := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		var name, ctype string
		var total float64
		if err := rows.Scan(&name, &ctype, &total); err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"category": name,
			"type":     ctype,
			"total":    total,
		})
	}

	// Check for any error that occurred during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
