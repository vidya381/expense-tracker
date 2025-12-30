package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vidya381/expense-tracker-backend/constants"
	"github.com/vidya381/expense-tracker-backend/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailExists        = errors.New("email_exists")
	ErrUsernameExists     = errors.New("username_exists")
	ErrUserNotFound       = errors.New("user_not_found")
	ErrInvalidCredentials = errors.New("invalid_credentials")
)

// RegisterUser creates a new user account with the provided credentials.
// Returns ErrEmailExists if email is already registered, ErrUsernameExists if username is taken.
// The password is hashed using bcrypt before storage.
func RegisterUser(ctx context.Context, db *sql.DB, username, email, password string) error {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	utils.LogInfo("RegisterUser started", "username", username, "email", email)

	// Check if email exists
	var exists bool
	utils.LogDebug("Checking if email exists", "email", email)
	err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		utils.LogError("Failed to check email existence", "error", err)
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		utils.LogInfo("Registration failed: email already exists", "email", email)
		return ErrEmailExists
	}
	utils.LogDebug("Email check passed", "email", email)

	// Check if username exists
	utils.LogDebug("Checking if username exists", "username", username)
	err = db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		utils.LogError("Failed to check username existence", "error", err)
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		utils.LogInfo("Registration failed: username already exists", "username", username)
		return ErrUsernameExists
	}
	utils.LogDebug("Username check passed", "username", username)

	// Hash the password
	utils.LogDebug("Hashing password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError("Password hashing failed", "error", err)
		return fmt.Errorf("password hashing failed: %w", err)
	}
	utils.LogDebug("Password hashed successfully")

	// Insert into users table
	utils.LogDebug("Inserting user into database", "username", username, "email", email)
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`
	_, err = db.ExecContext(ctx, query, username, email, string(hashedPassword))
	if err != nil {
		utils.LogError("Failed to insert user", "error", err, "username", username, "email", email)
		return fmt.Errorf("error inserting user: %w", err)
	}

	utils.LogInfo("User registered successfully", "username", username, "email", email)
	return nil
}

// LoginUser authenticates a user with email and password, returning a JWT token on success.
// Returns ErrUserNotFound if the email doesn't exist, ErrInvalidCredentials if password is incorrect.
// The JWT token expires after 72 hours and contains the user ID in its claims.
func LoginUser(ctx context.Context, db *sql.DB, email, password, jwtSecret string) (string, error) {
	ctx, cancel := utils.DBContext(ctx)
	defer cancel()

	utils.LogInfo("LoginUser started", "email", email)

	var userID int
	var hashedPassword string

	utils.LogDebug("Querying user from database", "email", email)
	err := db.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = $1", email).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		utils.LogInfo("Login failed: user not found", "email", email)
		return "", ErrUserNotFound
	}
	if err != nil {
		utils.LogError("Failed to query user by email", "error", err, "email", email)
		return "", fmt.Errorf("failed to query user by email: %w", err)
	}
	utils.LogDebug("User found in database", "email", email, "userID", userID)

	utils.LogDebug("Comparing password hash")
	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		utils.LogInfo("Login failed: invalid password", "email", email)
		return "", ErrInvalidCredentials
	}
	utils.LogDebug("Password verified successfully")

	// Create and sign JWT
	utils.LogDebug("Creating JWT token", "userID", userID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(constants.JWTExpirationHours).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		utils.LogError("Failed to sign JWT", "error", err, "userID", userID)
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	utils.LogInfo("User logged in successfully", "email", email, "userID", userID, "tokenLength", len(tokenString))
	return tokenString, nil
}
