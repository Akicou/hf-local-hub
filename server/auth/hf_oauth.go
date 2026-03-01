package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type HFProvider struct {
	config     *oauth2.Config
	auth       *Middleware
	stateStore map[string]string
}

func NewHFProvider(clientID, clientSecret, callbackURL string, auth *Middleware) *HFProvider {
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

	return &HFProvider{config: cfg, auth: auth, stateStore: make(map[string]string)}
}

func (p *HFProvider) Login(c *gin.Context) {
	state := generateState()
	p.stateStore[state] = "pending"

	url := p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (p *HFProvider) Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	if p.stateStore[state] != "pending" {
		c.JSON(400, gin.H{"error": "Invalid state"})
		return
	}
	delete(p.stateStore, state)

	token, err := p.config.Exchange(c.Request.Context(), code)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to exchange token"})
		return
	}

	userInfo, err := p.getUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get user info"})
		return
	}

	jwtToken, err := p.auth.GenerateToken(userInfo["sub"].(string), userInfo["name"].(string), "hf")
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(200, gin.H{"token": jwtToken, "user": userInfo})
}

func (p *HFProvider) getUserInfo(accessToken string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"sub":  "hf-user",
		"name": "HF User",
		"email": "user@huggingface.co",
	}, nil
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
