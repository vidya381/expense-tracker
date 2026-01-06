package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vidya381/myspendo-backend/utils"
)

// Key type for setting/retrieving user ID in context
type contextKey string

const userIDKey contextKey = "user_id"

// RequireAuth is a middleware that validates JWT tokens and extracts user ID.
// Protects routes by requiring a valid Bearer token in the Authorization header.
// The user ID from the token is stored in the request context for use by handlers.
func RequireAuth(jwtSecret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			utils.RespondWithUnauthorized(w, "Missing or invalid Authorization header")
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and verify JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure token signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			utils.RespondWithUnauthorized(w, "Invalid token")
			return
		}

		// Extract user ID from claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			utils.RespondWithUnauthorized(w, "Invalid token claims")
			return
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			utils.RespondWithUnauthorized(w, "Invalid user_id in token")
			return
		}
		userID := int(userIDFloat)

		// Pass user ID in context to the next handler
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// GetUserID retrieves the authenticated user's ID from the request context.
// Returns the user ID and true if found, or 0 and false if not authenticated.
// Should be called from handlers that are protected by RequireAuth middleware.
func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(userIDKey).(int)
	return userID, ok
}
