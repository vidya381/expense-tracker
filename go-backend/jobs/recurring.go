package jobs

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/vidya381/expense-tracker-backend/models"
	"github.com/vidya381/expense-tracker-backend/utils"
)

// Launches the recurring transaction processor in a background goroutine.
// Returns a channel that can be closed to stop the job gracefully.
func StartRecurringJob(db *sql.DB) chan struct{} {
	quit := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run once immediately on startup
		ProcessRecurringTransactions(db)

		for {
			select {
			case <-ticker.C:
				ProcessRecurringTransactions(db)
			case <-quit:
				fmt.Println("Recurring job shutting down gracefully...")
				return
			}
		}
	}()
	return quit
}

// Checks all recurring rules, schedules transactions as needed.
func ProcessRecurringTransactions(db *sql.DB) {
	// Use PostgreSQL advisory lock to prevent multiple instances from processing simultaneously
	// Lock ID: 123456789 (arbitrary number for this specific job)
	var lockAcquired bool
	err := db.QueryRow("SELECT pg_try_advisory_lock(123456789)").Scan(&lockAcquired)
	if err != nil {
		fmt.Println("Recurring jobs: error acquiring lock:", err)
		return
	}
	if !lockAcquired {
		// Another instance is already processing, skip this run
		return
	}

	// Ensure we release the lock when done
	defer func() {
		_, err := db.Exec("SELECT pg_advisory_unlock(123456789)")
		if err != nil {
			fmt.Println("Recurring jobs: error releasing lock:", err)
		}
	}()

	rows, err := db.Query(`
		SELECT id, user_id, category_id, amount, description, start_date, recurrence, last_occurrence
		FROM recurring_transactions
	`)
	if err != nil {
		fmt.Println("Recurring jobs: error querying:", err)
		return
	}
	defer rows.Close()
	now := time.Now().UTC().Truncate(24 * time.Hour)

	for rows.Next() {
		var rt models.RecurringTransaction
		var lastOccurrence sql.NullTime
		var startDate time.Time

		err := rows.Scan(&rt.ID, &rt.UserID, &rt.CategoryID, &rt.Amount, &rt.Description, &startDate, &rt.Recurrence, &lastOccurrence)
		if err != nil {
			fmt.Println("Recurring jobs: error scanning row:", err)
			continue
		}
		rt.StartDate = startDate.Format("2006-01-02")
		if lastOccurrence.Valid {
			rt.LastOccurrence = &lastOccurrence.Time
		} else {
			rt.LastOccurrence = nil
		}

		// fmt.Printf("Checking recurring id=%d desc=%q start=%s last_occurrence=%v recurrence=%s\n",
		// 	rt.ID, rt.Description, rt.StartDate, rt.LastOccurrence, rt.Recurrence)

		dueDates := GetAllMissedDueDates(rt, now)
		if len(dueDates) > 0 {
			ctx, cancel := utils.DBContext()
			for _, dueDate := range dueDates {
				_, err := db.ExecContext(ctx,
					`INSERT INTO transactions (user_id, category_id, amount, description, date)
					VALUES ($1, $2, $3, $4, $5)`,
					rt.UserID, rt.CategoryID, rt.Amount, rt.Description, dueDate.Format("2006-01-02"),
				)
				if err != nil {
					fmt.Println("Recurring jobs: error creating transaction:", err)
					continue
				}
				// fmt.Printf("Created recurring transaction instance for user %d on %s\n", rt.UserID, dueDate.Format("2006-01-02"))
			}
			// Update last_occurrence to latest due date
			latestDue := dueDates[len(dueDates)-1]
			_, err = db.ExecContext(ctx,
				`UPDATE recurring_transactions SET last_occurrence = $1 WHERE id = $2`,
				latestDue.Format("2006-01-02"), rt.ID)
			cancel()
			if err != nil {
				fmt.Println("Recurring jobs: error updating last_occurrence:", err)
			} else {
				fmt.Printf("Updated last_occurrence for recurring id=%d to %s\n", rt.ID, latestDue.Format("2006-01-02"))
			}
		}
	}
}

// Returns all the recurrence dates up to today (inclusive).
func GetAllMissedDueDates(rt models.RecurringTransaction, today time.Time) []time.Time {
	layout := "2006-01-02"
	start, err := time.Parse(layout, rt.StartDate)
	if err != nil {
		return nil
	}

	// Normalize to midnight UTC to avoid timezone issues
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	// Start date is in the future, no transactions due yet
	if start.After(today) {
		return nil
	}

	var next time.Time
	if rt.LastOccurrence != nil {
		last := *rt.LastOccurrence
		last = time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, time.UTC)
		// Calculate next occurrence after last
		next = calculateNextOccurrence(last, start, rt.Recurrence)
	} else {
		// No last occurrence, start from the start date
		next = start
	}

	var dueDates []time.Time
	// Limit to prevent infinite loops or excessive processing (max 3650 days / ~10 years of daily)
	maxIterations := 3650
	iterations := 0

	for !next.After(today) && iterations < maxIterations {
		dueDates = append(dueDates, next)
		next = calculateNextOccurrence(next, start, rt.Recurrence)
		iterations++
	}

	return dueDates
}

// calculateNextOccurrence calculates the next occurrence date based on recurrence type
// For monthly recurrence, preserves the original day-of-month from start date when possible
func calculateNextOccurrence(current time.Time, start time.Time, recurrence string) time.Time {
	switch recurrence {
	case "daily":
		return current.AddDate(0, 0, 1)
	case "weekly":
		return current.AddDate(0, 0, 7)
	case "monthly":
		// Preserve the day from start date, handling month-end edge cases
		targetDay := start.Day()
		next := current.AddDate(0, 1, 0)

		// Handle month-end dates (e.g., Jan 31 -> Feb 28/29)
		// Get last day of the target month
		firstOfNextMonth := time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		lastDayOfMonth := firstOfNextMonth.AddDate(0, 0, -1).Day()

		if targetDay > lastDayOfMonth {
			// Use last day of month if target day doesn't exist
			return time.Date(next.Year(), next.Month(), lastDayOfMonth, 0, 0, 0, 0, time.UTC)
		}
		return time.Date(next.Year(), next.Month(), targetDay, 0, 0, 0, 0, time.UTC)
	case "yearly":
		// Preserve month and day from start date, handling Feb 29 edge case
		targetMonth := start.Month()
		targetDay := start.Day()
		nextYear := current.Year() + 1

		// Handle Feb 29 on non-leap years
		if targetMonth == time.February && targetDay == 29 {
			// Check if next year is a leap year
			if !isLeapYear(nextYear) {
				// Use Feb 28 instead
				return time.Date(nextYear, time.February, 28, 0, 0, 0, 0, time.UTC)
			}
		}
		return time.Date(nextYear, targetMonth, targetDay, 0, 0, 0, 0, time.UTC)
	default:
		// Unknown recurrence, return current (will cause loop to exit)
		return current
	}
}

// isLeapYear checks if a year is a leap year
func isLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
