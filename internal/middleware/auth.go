package middleware

import (
	"net/http"
	"strings"

	"event-ticketing-system/internal/auth"
	"event-ticketing-system/internal/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// JWTAuth middleware validates JWT tokens
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))
		userRole := claims["role"].(string)

		// Get user from database to ensure they still exist
		db := c.MustGet("db").(*gorm.DB)
		var user models.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user info in context for handlers to use
		c.Set("user_id", userID)
		c.Set("user_role", userRole)
		c.Set("user", user)

		c.Next()
	}
}

// AdminAuth middleware ensures user has admin role
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		if userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}