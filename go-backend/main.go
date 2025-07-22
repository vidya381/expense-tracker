package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"

	"github.com/vidya381/expense-tracker-backend/handlers"
	"github.com/vidya381/expense-tracker-backend/jobs"
	"github.com/vidya381/expense-tracker-backend/middleware"
	"github.com/vidya381/expense-tracker-backend/models"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver with database/sql
	"github.com/joho/godotenv"
)

var jwtSecret = os.Getenv("JWT_SECRET")

var db *sql.DB

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found")
	}

	// Connect to database
	db, err = sql.Open("pgx", getDBConnURL())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Simple ping to verify connection is valid
	if err := db.Ping(); err != nil {
		panic("Failed to ping DB: " + err.Error())
	}

	fmt.Println("Connected to PostgreSQL successfully!")

	jobs.StartRecurringJob(db)

	// Define routes
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	// Protected routes (require JWT in Authorization header)
	http.HandleFunc("/category/add", middleware.RequireAuth(jwtSecret, addCategoryHandler))
	http.HandleFunc("/category/list", middleware.RequireAuth(jwtSecret, listCategoryHandler))
	http.HandleFunc("/transaction/add", middleware.RequireAuth(jwtSecret, addTransactionHandler))
	http.HandleFunc("/transaction/list", middleware.RequireAuth(jwtSecret, listTransactionHandler))
	http.HandleFunc("/transaction/update", middleware.RequireAuth(jwtSecret, updateTransactionHandler))
	http.HandleFunc("/transaction/delete", middleware.RequireAuth(jwtSecret, deleteTransactionHandler))
	http.HandleFunc("/summary/totals", middleware.RequireAuth(jwtSecret, summaryTotalsHandler))
	http.HandleFunc("/summary/monthly", middleware.RequireAuth(jwtSecret, summaryMonthlyHandler))
	http.HandleFunc("/summary/category", middleware.RequireAuth(jwtSecret, summaryCategoryHandler))
	http.HandleFunc("/summary/group", middleware.RequireAuth(jwtSecret, summaryGroupHandler))
	http.HandleFunc("/summary/category/monthly", middleware.RequireAuth(jwtSecret, summaryCategoryMonthHandler))
	http.HandleFunc("/export", middleware.RequireAuth(jwtSecret, exportTransactionsHandler))
	http.HandleFunc("/recurring/add", middleware.RequireAuth(jwtSecret, addRecurringHandler))
	http.HandleFunc("/recurring/list", middleware.RequireAuth(jwtSecret, listRecurringHandler))
	http.HandleFunc("/recurring/edit", middleware.RequireAuth(jwtSecret, editRecurringHandler))
	http.HandleFunc("/recurring/delete", middleware.RequireAuth(jwtSecret, deleteRecurringHandler))
	http.HandleFunc("/transactions/search", middleware.RequireAuth(jwtSecret, searchAndFilterTransactionsHandler))

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// Builds the PostgreSQL connection URL from environment variables for use with sql.Open
func getDBConnURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

// Handles user registration via POST request (expects 'username', 'email', 'password')
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	// Basic validation
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || !strings.Contains(parts[1], ".") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	if len(password) < 4 {
		http.Error(w, "Password must be at least 4 characters", http.StatusBadRequest)
		return
	}

	err := handlers.RegisterUser(db, username, email, password)
	if err != nil {
		http.Error(w, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User registered successfully!"))
}

// Handles user login via POST request (expects 'email', 'password')
// Returns a JWT token if credentials are valid
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	token, err := handlers.LoginUser(db, email, password, jwtSecret)
	if err != nil {
		http.Error(w, "Login failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Write([]byte(token))
}

// Handles category creation for a user (expects 'user_id', 'name', 'type' in POST form)
func addCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	ctype := strings.ToLower(strings.TrimSpace(r.FormValue("type")))

	if name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}
	if ctype != "expense" && ctype != "income" {
		http.Error(w, "Category type must be 'expense' or 'income'", http.StatusBadRequest)
		return
	}

	err := handlers.AddCategory(db, userID, name, ctype)
	if err != nil {
		http.Error(w, "Add category failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte("Category added successfully"))
}

// Handles listing all categories for a user (expects 'user_id' as a URL query parameter)
func listCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}
	cats, err := handlers.ListCategories(db, userID)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}

// Add a new transaction via POST
func addTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil || categoryID <= 0 {
		http.Error(w, "Valid category_id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Amount must be a number greater than zero", http.StatusBadRequest)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	date := r.FormValue("date")
	if date == "" {
		http.Error(w, "Transaction date is required", http.StatusBadRequest)
		return
	}

	tx := models.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	err = handlers.AddTransaction(db, tx)
	if err != nil {
		http.Error(w, "Error adding transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Transaction added successfully"))
}

// List all transactions for a user (GET)
func listTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}
	list, err := handlers.ListTransactions(db, userID)
	if err != nil {
		http.Error(w, "Failed to fetch transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

// Update an existing transaction (POST)
func updateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Valid transaction ID is required", http.StatusBadRequest)
		return
	}
	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil || categoryID <= 0 {
		http.Error(w, "Valid category_id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Amount must be a number greater than zero", http.StatusBadRequest)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	date := r.FormValue("date")
	if date == "" {
		http.Error(w, "Transaction date is required", http.StatusBadRequest)
		return
	}

	tx := models.Transaction{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	err = handlers.UpdateTransaction(db, tx)
	if err != nil {
		http.Error(w, "Error updating transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Transaction updated successfully"))
}

// Delete a transaction (POST)
func deleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil || id <= 0 {
		http.Error(w, "Valid transaction ID is required", http.StatusBadRequest)
		return
	}

	err = handlers.DeleteTransaction(db, id, userID)
	if err != nil {
		http.Error(w, "Failed to delete transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Transaction deleted"))
}

// Returns overall totals for this user
func summaryTotalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	expenses, income, err := handlers.GetTotals(db, userID)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]float64{
		"total_expenses": expenses,
		"total_income":   income,
	})
}

// Returns monthly group totals for this user
func summaryMonthlyHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	summary, err := handlers.GetMonthlyTotals(db, userID)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

// Returns category-wise breakdown for this user for a date range
func summaryCategoryHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	result, err := handlers.GetCategoryBreakdown(db, userID, from, to)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// Group by period endpoint
func summaryGroupHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	granularity := r.URL.Query().Get("by")
	summary, err := handlers.GetGroupTotals(db, userID, granularity)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(summary)
}

// Category-wise summary for given year/month
func summaryCategoryMonthHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 || month == 0 {
		http.Error(w, "year and month required", http.StatusBadRequest)
		return
	}
	result, err := handlers.GetCategoryMonthSummary(db, userID, year, month)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(result)
}

// Export transactions to CSV or JSON
func exportTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	transactions, err := handlers.ListTransactions(db, userID)
	if err != nil {
		http.Error(w, "Failed to fetch transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	format := r.URL.Query().Get("format")
	if format == "csv" {
		w.Header().Set("Content-Disposition", "attachment;filename=transactions.csv")
		w.Header().Set("Content-Type", "text/csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"ID", "CategoryID", "Amount", "Description", "Date", "CreatedAt"})
		for _, tx := range transactions {
			writer.Write([]string{
				strconv.Itoa(tx.ID),
				strconv.Itoa(tx.CategoryID),
				fmt.Sprintf("%.2f", tx.Amount),
				tx.Description,
				tx.Date,
				tx.CreatedAt,
			})
		}
		writer.Flush()
		return
	}
	// Default: JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// User to add a recurring transaction.
func addRecurringHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
	if err != nil || categoryID <= 0 {
		http.Error(w, "Valid category_id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Amount must be a positive number", http.StatusBadRequest)
		return
	}
	description := strings.TrimSpace(r.FormValue("description"))
	startDate := r.FormValue("start_date")
	recurrence := strings.ToLower(strings.TrimSpace(r.FormValue("recurrence")))
	if startDate == "" || recurrence == "" {
		http.Error(w, "start_date and recurrence are required", http.StatusBadRequest)
		return
	}

	rt := models.RecurringTransaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		StartDate:   startDate,
		Recurrence:  recurrence,
	}

	err = handlers.AddRecurringTransaction(db, rt)
	if err != nil {
		http.Error(w, "Failed to add recurring transaction: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Recurring transaction added!"))
}

// Returns all recurring transactions for the authenticated user
func listRecurringHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	recurrings, err := handlers.ListRecurringTransactions(db, userID)
	if err != nil {
		http.Error(w, "Failed to list recurring transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(recurrings)
}

func editRecurringHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Valid id is required", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseFloat(r.FormValue("amount"), 64)
	if err != nil || amount <= 0 {
		http.Error(w, "Amount must be a positive number", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	startDate := r.FormValue("start_date")
	recurrence := strings.ToLower(strings.TrimSpace(r.FormValue("recurrence")))

	err = handlers.EditRecurringTransaction(db, userID, id, amount, description, startDate, recurrence)
	if err != nil {
		http.Error(w, "Failed to edit recurring transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Recurring transaction updated!"))
}

func deleteRecurringHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := middleware.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, "Valid id is required", http.StatusBadRequest)
		return
	}
	err = handlers.DeleteRecurringTransaction(db, id, userID)
	if err != nil {
		http.Error(w, "Failed to delete recurring transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Recurring transaction deleted!"))
}

func searchAndFilterTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		jsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Pagination
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	// Sorting
	sortParam := r.URL.Query().Get("sort")
	allowedSorts := map[string]string{
		"date_asc":    "date ASC",
		"date_desc":   "date DESC",
		"amount_asc":  "amount ASC",
		"amount_desc": "amount DESC",
	}
	orderBy := "date DESC"
	if s, ok := allowedSorts[sortParam]; ok {
		orderBy = s
	}

	// Filters
	keyword := r.URL.Query().Get("q")
	categoryID, _ := strconv.Atoi(r.URL.Query().Get("category_id"))
	dateFrom := r.URL.Query().Get("from")
	dateTo := r.URL.Query().Get("to")
	amountMin, _ := strconv.ParseFloat(r.URL.Query().Get("min_amount"), 64)
	amountMax, _ := strconv.ParseFloat(r.URL.Query().Get("max_amount"), 64)

	list, err := handlers.FilterTransactionsPaginated(
		db, userID, keyword, categoryID, dateFrom, dateTo, amountMin, amountMax, orderBy, limit, offset,
	)
	if err != nil {
		jsonError(w, "Search error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
