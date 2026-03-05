package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Akicou/hf-local-hub/server/db"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	jwt.RegisteredClaims
}

type Middleware struct {
	jwtSecret string
	db        *gorm.DB
}

func NewMiddleware(jwtSecret string, database *gorm.DB) *Middleware {
	return &Middleware{
		jwtSecret: jwtSecret,
		db:        database,
	}
}

// Required middleware requires authentication (JWT or API token)
func (m *Middleware) Required() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		// Try to extract token
		token, err := extractToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// First try JWT authentication
		claims, err := m.validateJWTToken(token)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("provider", claims.Provider)
			c.Set("auth_type", "jwt")
			c.Next()
			return
		}

		// If JWT fails, try API token authentication
		apiToken, perms, err := m.validateAPIToken(token)
		if err == nil {
			c.Set("user_id", apiToken.UserID)
			c.Set("username", apiToken.Name)
			c.Set("provider", "api_token")
			c.Set("auth_type", "api_token")
			c.Set("token_permissions", perms)
			c.Next()
			return
		}

		c.JSON(401, gin.H{"error": "Invalid token"})
		c.Abort()
	}
}

// Optional middleware attempts authentication but doesn't require it
func (m *Middleware) Optional() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		token, err := extractToken(authHeader)
		if err != nil {
			c.Next()
			return
		}

		// Try JWT first
		claims, err := m.validateJWTToken(token)
		if err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("provider", claims.Provider)
			c.Set("auth_type", "jwt")
			c.Next()
			return
		}

		// Try API token
		apiToken, perms, err := m.validateAPIToken(token)
		if err == nil {
			c.Set("user_id", apiToken.UserID)
			c.Set("username", apiToken.Name)
			c.Set("provider", "api_token")
			c.Set("auth_type", "api_token")
			c.Set("token_permissions", perms)
		}

		c.Next()
	}
}

// RequirePermission middleware checks if the user has a specific permission
func (m *Middleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If authenticated via JWT, check if user is admin or has full permissions
		authType, exists := c.Get("auth_type")
		if exists && authType == "jwt" {
			// JWT users have full permissions (authenticated via OAuth/LDAP)
			c.Next()
			return
		}

		// If authenticated via API token, check permissions
		permsInterface, exists := c.Get("token_permissions")
		if !exists {
			c.JSON(403, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		perms, ok := permsInterface.(db.TokenPermissions)
		if !ok {
			c.JSON(403, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		// Check the specific permission
		hasPermission := false
		switch permission {
		case "read":
			hasPermission = perms.Read
		case "write":
			hasPermission = perms.Write
		case "delete":
			hasPermission = perms.Delete
		case "admin":
			hasPermission = perms.Admin
		}

		if !hasPermission {
			c.JSON(403, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GenerateToken generates a JWT token for a user
func (m *Middleware) GenerateToken(userID, username, provider string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Provider: provider,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

// RefreshToken refreshes a JWT token
func (m *Middleware) RefreshToken(tokenString string) (string, error) {
	claims, err := m.validateJWTToken(tokenString)
	if err != nil {
		return "", err
	}
	return m.GenerateToken(claims.UserID, claims.Username, claims.Provider)
}

// GenerateAPIToken generates a new API token for a user
func (m *Middleware) GenerateAPIToken(userID, name string, permissions db.TokenPermissions, expiresAt *time.Time) (*db.APIToken, error) {
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := "hf_" + hex.EncodeToString(tokenBytes)

	// Serialize permissions to JSON
	permsJSON, err := json.Marshal(permissions)
	if err != nil {
		return nil, err
	}

	apiToken := &db.APIToken{
		Token:       token,
		Name:        name,
		UserID:      userID,
		Permissions: string(permsJSON),
		ExpiresAt:   expiresAt,
	}

	if err := m.db.Create(apiToken).Error; err != nil {
		return nil, err
	}

	return apiToken, nil
}

// ListAPITokens lists all API tokens for a user
func (m *Middleware) ListAPITokens(userID string) ([]db.APIToken, error) {
	var tokens []db.APIToken
	if err := m.db.Where("user_id = ?", userID).Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// DeleteAPIToken deletes an API token
func (m *Middleware) DeleteAPIToken(userID, tokenID string) error {
	result := m.db.Where("id = ? AND user_id = ?", tokenID, userID).Delete(&db.APIToken{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token not found")
	}
	return nil
}

func extractToken(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization format")
	}
	return parts[1], nil
}

func (m *Middleware) validateJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (m *Middleware) validateAPIToken(tokenString string) (*db.APIToken, *db.TokenPermissions, error) {
	var apiToken db.APIToken
	if err := m.db.Where("token = ?", tokenString).First(&apiToken).Error; err != nil {
		return nil, nil, errors.New("invalid API token")
	}

	// Check if expired
	if apiToken.IsExpired() {
		return nil, nil, errors.New("API token expired")
	}

	// Parse permissions
	var perms db.TokenPermissions
	if err := json.Unmarshal([]byte(apiToken.Permissions), &perms); err != nil {
		return nil, nil, errors.New("invalid permissions format")
	}

	// Update last used timestamp
	now := time.Now()
	apiToken.LastUsedAt = &now
	m.db.Model(&apiToken).Update("last_used_at", now)

	return &apiToken, &perms, nil
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetUsername extracts username from gin context
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}
