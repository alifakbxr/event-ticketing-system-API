package auth

import (
	"errors"
	"time"

	"event-ticketing-system/internal/models"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your-secret-key-change-this-in-production")

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

// GenerateToken generates a JWT token for a user
func GenerateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := &Claims{
		UserID: user.ID,
		Role:   user.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token string
func ValidateToken(tokenString string) (*jwt.Token, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword verifies a password against its hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}