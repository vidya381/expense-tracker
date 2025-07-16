package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(db *sql.DB, username, email, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	// Insert into users table
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`
	_, err = db.ExecContext(context.Background(), query, username, email, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}

func LoginUser(db *sql.DB, email, password, jwtSecret string) (string, error) {
	var userID int
	var hashedPassword string

	err := db.QueryRowContext(context.Background(),
		"SELECT id, password FROM users WHERE email = $1", email).Scan(&userID, &hashedPassword)
	if err != nil {
		return "", fmt.Errorf("user not found or db error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	// Create and sign JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return tokenString, nil
}
