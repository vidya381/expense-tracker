package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/vidya381/expense-tracker-backend/handlers"
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

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

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

	email := r.FormValue("email")
	password := r.FormValue("password")

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
	name := r.FormValue("name")
	ctype := r.FormValue("type")

	err := handlers.AddCategory(db, userID, name, ctype)
	if err != nil {
		http.Error(w, "Add category failed: "+err.Error(), http.StatusInternalServerError)
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
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	description := r.FormValue("description")
	date := r.FormValue("date")

	tx := models.Transaction{
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	err := handlers.AddTransaction(db, tx)
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
	id, _ := strconv.Atoi(r.FormValue("id"))
	categoryID, _ := strconv.Atoi(r.FormValue("category_id"))
	amount, _ := strconv.ParseFloat(r.FormValue("amount"), 64)
	description := r.FormValue("description")
	date := r.FormValue("date")

	tx := models.Transaction{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Description: description,
		Date:        date,
	}

	err := handlers.UpdateTransaction(db, tx)
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
	id, _ := strconv.Atoi(r.FormValue("id"))

	err := handlers.DeleteTransaction(db, id, userID)
	if err != nil {
		http.Error(w, "Failed to delete transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Transaction deleted"))
}
