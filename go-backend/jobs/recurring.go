package jobs

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/vidya381/myspendo-backend/models"
	"github.com/vidya381/myspendo-backend/utils"
)

// StartRecurringJob launches the recurring transaction processor in a background goroutine.
// Processes recurring transactions every hour to generate actual transactions based on schedules.
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
				slog.Info("Recurring job shutting down gracefully")
				return
			}
		}
	}()
	return quit
}

// ProcessRecurringTransactions checks all recurring transaction rules and creates due transactions.
// Uses PostgreSQL advisory locks to prevent concurrent processing by multiple instances.
// Transactions are created for all missed occurrences up to the current date.
func ProcessRecurringTransactions(db *sql.DB) {
	// Use PostgreSQL advisory lock to prevent multiple instances from processing simultaneously
	// Lock ID: 123456789 (arbitrary number for this specific job)
	var lockAcquired bool
	err := db.QueryRow("SELECT pg_try_advisory_lock(123456789)").Scan(&lockAcquired)
	if err != nil {
		slog.Error("Recurring jobs: error acquiring lock", "error", err)
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
			slog.Error("Recurring jobs: error releasing lock", "error", err)
		}
	}()

	rows, err := db.Query(`
		SELECT id, user_id, category_id, amount, description, start_date, recurrence, last_occurrence
		FROM recurring_transactions
	`)
	if err != nil {
		slog.Error("Recurring jobs: error querying", "error", err)
		return
	}
	defer rows.Close()
	// Use UTC and truncate to midnight for consistent date-only comparison across timezones
	now := time.Now().UTC().Truncate(24 * time.Hour)

	var totalProcessed, totalFailed, totalCreated int

	for rows.Next() {
		var rt models.RecurringTransaction
		var lastOccurrence sql.NullTime
		var startDate time.Time

		err := rows.Scan(&rt.ID, &rt.UserID, &rt.CategoryID, &rt.Amount, &rt.Description, &startDate, &rt.Recurrence, &lastOccurrence)
		if err != nil {
			slog.Error("Recurring jobs: error scanning row", "error", err)
			continue
		}
		rt.StartDate = startDate.Format("2006-01-02")
		if lastOccurrence.Valid {
			rt.LastOccurrence = &lastOccurrence.Time
		} else {
			rt.LastOccurrence = nil
		}

		dueDates := GetAllMissedDueDates(rt, now)
		if len(dueDates) > 0 {
			totalProcessed++
			var failedCount int
			// Create transactions with fresh context for each batch
			for _, dueDate := range dueDates {
				ctx, cancel := utils.DBContext(nil)
				_, err := db.ExecContext(ctx,
					`INSERT INTO transactions (user_id, category_id, amount, description, date)
					VALUES ($1, $2, $3, $4, $5)`,
					rt.UserID, rt.CategoryID, rt.Amount, rt.Description, dueDate.Format("2006-01-02"),
				)
				cancel()
				if err != nil {
					slog.Error("Recurring jobs: error creating transaction", "error", err, "recurring_id", rt.ID, "date", dueDate.Format("2006-01-02"))
					failedCount++
					continue
				}
				totalCreated++
			}

			if failedCount > 0 {
				totalFailed++
			}

			// Update last_occurrence to latest due date with fresh context
			latestDue := dueDates[len(dueDates)-1]
			ctx, cancel := utils.DBContext(nil)
			_, err = db.ExecContext(ctx,
				`UPDATE recurring_transactions SET last_occurrence = $1 WHERE id = $2`,
				latestDue.Format("2006-01-02"), rt.ID)
			cancel()
			if err != nil {
				slog.Error("Recurring jobs: error updating last_occurrence", "error", err, "recurring_id", rt.ID)
			} else {
				slog.Info("Updated recurring transaction", "recurring_id", rt.ID, "last_occurrence", latestDue.Format("2006-01-02"))
			}
		}
	}

	// Log summary of job execution
	if totalProcessed > 0 {
		slog.Info("Recurring job completed",
			"processed", totalProcessed,
			"created", totalCreated,
			"failed_rules", totalFailed)
	}

	if totalFailed > 0 {
		slog.Warn("Recurring job had failures",
			"failed_rules", totalFailed,
			"total_processed", totalProcessed)
	}
}

// GetAllMissedDueDates calculates all due dates for a recurring transaction up to today (inclusive).
// Handles month-end and leap year edge cases for monthly and yearly recurrences.
// Returns an empty slice if the start date is in the future or if there are no due dates.
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
