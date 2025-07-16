package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/vidya381/expense-tracker-backend/handlers"

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
	http.HandleFunc("/category/add", addCategoryHandler)
	http.HandleFunc("/category/list", listCategoryHandler)

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
	userIDStr := r.FormValue("user_id")
	name := r.FormValue("name")
	ctype := r.FormValue("type")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}

	err = handlers.AddCategory(db, userID, name, ctype)
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
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user_id", http.StatusBadRequest)
		return
	}
	cats, err := handlers.ListCategories(db, userID)
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cats)
}
