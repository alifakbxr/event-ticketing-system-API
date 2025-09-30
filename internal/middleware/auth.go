package middleware

import (
	"context"
	"net/http"
	"strings"

	"event-ticketing-system/internal/auth"
	"event-ticketing-system/internal/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
)

// JWTAuth middleware validates JWT tokens
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader {
			http.Error(w, `{"error": "Bearer token required"}`, http.StatusUnauthorized)
			return
		}

		// Parse and validate token
		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Set user information in context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		userID := uint(claims["user_id"].(float64))
		userRole := claims["role"].(string)

		// Get user from database to ensure they still exist
		db := r.Context().Value("db").(*gorm.DB)
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusUnauthorized)
			return
		}

		// Set user info in context for handlers to use
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user_role", userRole)
		ctx = context.WithValue(ctx, "user", user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminAuth middleware ensures user has admin role
func AdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value("user_role")
		if userRole == nil {
			http.Error(w, `{"error": "User role not found"}`, http.StatusUnauthorized)
			return
		}

		if userRole != "admin" {
			http.Error(w, `{"error": "Admin access required"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}