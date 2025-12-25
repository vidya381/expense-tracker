package handlers

import (
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

func RegisterUser(db *sql.DB, username, email, password string) error {
	// Check if email exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrEmailExists
	}

	// Check if username exists
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Insert into users table
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`
	ctx, cancel := utils.DBContext()
	defer cancel()
	_, err = db.ExecContext(ctx, query, username, email, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}

func LoginUser(db *sql.DB, email, password, jwtSecret string) (string, error) {
	var userID int
	var hashedPassword string

	err := db.QueryRow("SELECT id, password FROM users WHERE email = $1", email).Scan(&userID, &hashedPassword)
	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	// Create and sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(constants.JWTExpirationHours).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}
