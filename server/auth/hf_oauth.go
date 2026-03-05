package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/Akicou/hf-local-hub/server/db"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type HFProvider struct {
	config   *oauth2.Config
	auth     *Middleware
	db       *gorm.DB
	logger   *zap.Logger
}

// HFUserInfo represents the response from Hugging Face /api/whoami-v2 endpoint
type HFUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatarUrl"`
	FullName string `json:"fullname"`
	Orgs     []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"orgs"`
}

// NewHFProvider creates a new Hugging Face OAuth provider with persistent state storage
func NewHFProvider(clientID, clientSecret, callbackURL string, auth *Middleware, database *gorm.DB, logger *zap.Logger) *HFProvider {
	cfg := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://huggingface.co/oauth/authorize",
			TokenURL: "https://huggingface.co/oauth/token",
		},
		Scopes: []string{"openid", "profile", "email"},
	}

	provider := &HFProvider{
		config: cfg,
		auth:   auth,
		db:     database,
		logger: logger,
	}

	// Clean up expired states periodically
	go provider.cleanupExpiredStates()

	return provider
}

// cleanupExpiredStates removes expired OAuth states from the database
func (p *HFProvider) cleanupExpiredStates() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		result := p.db.Where("expires_at < ? OR status = ?", time.Now(), "used").Delete(&db.OAuthState{})
		if result.Error != nil {
			p.logger.Error("Failed to cleanup expired OAuth states", zap.Error(result.Error))
		} else if result.RowsAffected > 0 {
			p.logger.Info("Cleaned up expired OAuth states", zap.Int64("count", result.RowsAffected))
		}
	}
}

// Login initiates the OAuth flow
func (p *HFProvider) Login(c *gin.Context) {
	state := generateState()

	// Store state in database with 10-minute expiration
	oauthState := &db.OAuthState{
		State:     state,
		Provider:  "hf",
		Status:    "pending",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := p.db.Create(oauthState).Error; err != nil {
		p.logger.Error("Failed to create OAuth state", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate login"})
		return
	}

	url := p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Callback handles the OAuth callback
func (p *HFProvider) Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	// Verify state in database
	var oauthState db.OAuthState
	if err := p.db.Where("state = ? AND provider = ? AND status = ?", state, "hf", "pending").First(&oauthState).Error; err != nil {
		p.logger.Warn("Invalid or expired OAuth state", zap.String("state", state))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired state"})
		return
	}

	// Check if expired
	if oauthState.IsExpired() {
		oauthState.Status = "expired"
		p.db.Save(&oauthState)
		c.JSON(http.StatusBadRequest, gin.H{"error": "State expired"})
		return
	}

	// Mark state as used
	oauthState.Status = "used"
	if err := p.db.Save(&oauthState).Error; err != nil {
		p.logger.Error("Failed to update OAuth state", zap.Error(err))
	}

	token, err := p.config.Exchange(c.Request.Context(), code)
	if err != nil {
		p.logger.Error("Failed to exchange token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	userInfo, err := p.getUserInfo(token.AccessToken)
	if err != nil {
		p.logger.Error("Failed to get user info", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	userID := userInfo["sub"].(string)
	username := userInfo["name"].(string)

	// Create or update user in database
	p.ensureUserExists(userID, username, userInfo["email"].(string), "hf")

	jwtToken, err := p.auth.GenerateToken(userID, username, "hf")
	if err != nil {
		p.logger.Error("Failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	p.logger.Info("User logged in via HF OAuth", zap.String("user_id", userID))

	c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": userInfo})
}

// ensureUserExists creates a user record if it doesn't exist
func (p *HFProvider) ensureUserExists(userID, username, email, provider string) {
	var user db.User
	if err := p.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		// User doesn't exist, create one
		user = db.User{
			UserID:   userID,
			Username: username,
			Email:    email,
			Provider: provider,
			IsActive: true,
		}
		if err := p.db.Create(&user).Error; err != nil {
			p.logger.Error("Failed to create user", zap.Error(err))
		} else {
			p.logger.Info("Created new user", zap.String("user_id", userID), zap.String("provider", provider))
		}
	} else {
		// Update user info if needed
		needsUpdate := false
		if user.Username != username {
			user.Username = username
			needsUpdate = true
		}
		if user.Email != email && email != "" {
			user.Email = email
			needsUpdate = true
		}
		if needsUpdate {
			p.db.Save(&user)
		}
	}
}

func (p *HFProvider) getUserInfo(accessToken string) (map[string]interface{}, error) {
	// Fetch user info from Hugging Face API
	req, err := http.NewRequest("GET", "https://huggingface.co/api/whoami-v2", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", "hf-local-hub/0.2.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Hugging Face API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var hfUser HFUserInfo
	if err := json.Unmarshal(body, &hfUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	// Use ID if available, otherwise use username
	userID := hfUser.ID
	if userID == "" {
		userID = hfUser.Username
	}

	// Use fullname if available, otherwise use username
	displayName := hfUser.FullName
	if displayName == "" {
		displayName = hfUser.Username
	}

	return map[string]interface{}{
		"sub":    userID,
		"name":   displayName,
		"email":  hfUser.Email,
		"id":     userID,
		"avatar": hfUser.Avatar,
	}, nil
}

func generateState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fallback-state-" + fmt.Sprint(time.Now().Unix())
	}
	return base64.URLEncoding.EncodeToString(b)
}
