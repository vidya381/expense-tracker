package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Returns the total expenses and income for the authenticated user.
func GetTotals(db *sql.DB, userID int) (expenses float64, income float64, err error) {
	err = db.QueryRowContext(context.Background(),
		`SELECT
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END),0)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1`, userID).Scan(&expenses, &income)
	return
}

// Returns monthly totals (income and expenses per month)
func GetMonthlyTotals(db *sql.DB, userID int) ([]map[string]interface{}, error) {
	rows, err := db.QueryContext(context.Background(),
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

	var results []map[string]interface{}
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
	return results, nil
}

// Returns category-wise breakdown for the user
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

	rows, err := db.QueryContext(context.Background(), base, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []map[string]interface{}
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
	return result, nil
}

// Group by month, week, or year for totals by period.
func GetGroupTotals(db *sql.DB, userID int, granularity string) ([]map[string]interface{}, error) {
	allowed := map[string]bool{"month": true, "week": true, "year": true}
	unit := "month"
	if allowed[granularity] {
		unit = granularity
	}
	sqlQuery := fmt.Sprintf(`
		SELECT DATE_TRUNC('%s', t.date) as period,
			COALESCE(SUM(CASE WHEN c.type = 'expense' THEN t.amount ELSE 0 END), 0) as total_expenses,
			COALESCE(SUM(CASE WHEN c.type = 'income' THEN t.amount ELSE 0 END), 0) as total_income
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1
		GROUP BY period ORDER BY period DESC`, unit)
	rows, err := db.QueryContext(context.Background(), sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
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
	return results, nil
}

// Category-wise summary for a specific month (or any period you want)
func GetCategoryMonthSummary(db *sql.DB, userID int, year, month int) ([]map[string]interface{}, error) {
	query := `
		SELECT c.name, c.type, COALESCE(SUM(t.amount),0)
		FROM transactions t
		JOIN categories c ON t.category_id = c.id
		WHERE t.user_id = $1 AND EXTRACT(YEAR FROM t.date) = $2 AND EXTRACT(MONTH FROM t.date) = $3
		GROUP BY c.name, c.type
		ORDER BY c.type, SUM(t.amount) DESC`
	rows, err := db.QueryContext(context.Background(), query, userID, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]interface{}
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
	return result, nil
}
