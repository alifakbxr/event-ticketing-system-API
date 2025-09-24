package handlers

import (
	"net/http"

	"event-ticketing-system/internal/auth"
	"event-ticketing-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)


// AuthHandler handles authentication related requests
type AuthHandler struct {
	db *gorm.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Register handles user registration
// swagger:operation POST /api/register auth register
// ---
// summary: Register a new user
// description: Creates a new user account with the provided information
// tags:
// - Authentication
// parameters:
// - name: request
//   in: body
//   description: User registration data
//   required: true
//   schema:
//     "$ref": "#/definitions/RegisterRequest"
// responses:
//   201:
//     description: User registered successfully
//     schema:
//       "$ref": "#/definitions/AuthResponse"
//   400:
//     description: Invalid request data
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   409:
//     description: User already exists
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists with this email"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user", // Default role
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// swagger:operation POST /api/login auth login
// ---
// summary: User login
// description: Authenticates a user and returns a JWT token
// tags:
// - Authentication
// parameters:
// - name: request
//   in: body
//   description: User login credentials
//   required: true
//   schema:
//     "$ref": "#/definitions/LoginRequest"
// responses:
//   200:
//     description: Login successful
//     schema:
//       "$ref": "#/definitions/AuthResponse"
//   400:
//     description: Invalid request data
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   401:
//     description: Invalid credentials
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Remove password from response
	user.Password = ""

	response := AuthResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout
// swagger:operation POST /api/logout auth logout
// ---
// summary: User logout
// description: Logs out the current user (client-side token removal)
// tags:
// - Authentication
// security:
// - Bearer: []
// responses:
//   200:
//     description: Logout successful
//     schema:
//       type: object
//       properties:
//         message:
//           type: string
//           example: "Logged out successfully"
//   401:
//     description: Unauthorized
//     schema:
//       "$ref": "#/definitions/ErrorResponse"
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT implementation, logout is typically handled on the client side
	// by removing the token. However, we can implement token blacklisting here if needed.

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}