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
	var last time.Time
	if rt.LastOccurrence != nil {
		last = *rt.LastOccurrence
	} else {
		last = start.AddDate(0, 0, -1)
	}

	var dueDates []time.Time
	next := last
	for {
		// Finding next recurrence date
		switch rt.Recurrence {
		case "daily":
			next = next.AddDate(0, 0, 1)
		case "weekly":
			next = next.AddDate(0, 0, 7)
		case "monthly":
			next = next.AddDate(0, 1, 0)
		case "yearly":
			next = next.AddDate(1, 0, 0)
		default:
			return dueDates
		}
		if next.After(today) {
			break
		}
		if !next.Before(start) {
			dueDates = append(dueDates, next)
		}
	}
	return dueDates
}
