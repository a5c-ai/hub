package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
)

var (
	ErrOAuthProviderNotConfigured = errors.New("oauth provider not configured")
	ErrInvalidOAuthState          = errors.New("invalid oauth state")
	ErrOAuthCodeExchange          = errors.New("oauth code exchange failed")
	ErrOAuthUserInfo              = errors.New("failed to get oauth user info")
)

type OAuthProvider interface {
	GetAuthURL(state string, redirectURI string) string
	ExchangeCode(code, redirectURI string) (*OAuthToken, error)
	GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error)
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type OAuthUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"login,omitempty"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar_url,omitempty"`
}

type OAuthManager struct {
	providers map[string]OAuthProvider
	config    *config.Config
}

func NewOAuthManager(cfg *config.Config) *OAuthManager {
	providers := make(map[string]OAuthProvider)
	
	// Initialize GitHub provider if configured
	if cfg.OAuth.GitHub.ClientID != "" && cfg.OAuth.GitHub.ClientSecret != "" {
		providers["github"] = &GitHubProvider{
			ClientID:     cfg.OAuth.GitHub.ClientID,
			ClientSecret: cfg.OAuth.GitHub.ClientSecret,
		}
	}
	
	// Initialize Google provider if configured
	if cfg.OAuth.Google.ClientID != "" && cfg.OAuth.Google.ClientSecret != "" {
		providers["google"] = &GoogleProvider{
			ClientID:     cfg.OAuth.Google.ClientID,
			ClientSecret: cfg.OAuth.Google.ClientSecret,
		}
	}

	return &OAuthManager{
		providers: providers,
		config:    cfg,
	}
}

func (m *OAuthManager) GetProvider(name string) (OAuthProvider, error) {
	provider, exists := m.providers[name]
	if !exists {
		return nil, ErrOAuthProviderNotConfigured
	}
	return provider, nil
}

func (m *OAuthManager) GenerateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GitHub OAuth Provider
type GitHubProvider struct {
	ClientID     string
	ClientSecret string
}

func (p *GitHubProvider) GetAuthURL(state, redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", p.ClientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("scope", "user:email")
	params.Add("state", state)
	
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *GitHubProvider) ExchangeCode(code, redirectURI string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthCodeExchange
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (p *GitHubProvider) GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthUserInfo
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo OAuthUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	// GitHub doesn't always return public email, fetch it separately
	if userInfo.Email == "" {
		if email, err := p.getUserEmail(token.AccessToken); err == nil {
			userInfo.Email = email
		}
	}

	return &userInfo, nil
}

func (p *GitHubProvider) getUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch user emails")
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

// Google OAuth Provider
type GoogleProvider struct {
	ClientID     string
	ClientSecret string
}

func (p *GoogleProvider) GetAuthURL(state, redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", p.ClientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *GoogleProvider) ExchangeCode(code, redirectURI string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthCodeExchange
	}

	var token OAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func (p *GoogleProvider) GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthUserInfo
	}

	var userInfo OAuthUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// OAuth service methods for AuthService
func (s *AuthService) InitiateOAuth(provider, redirectURI string) (string, string, error) {
	oauthManager := NewOAuthManager(s.config)
	
	oauthProvider, err := oauthManager.GetProvider(provider)
	if err != nil {
		return "", "", err
	}

	state, err := oauthManager.GenerateState()
	if err != nil {
		return "", "", err
	}

	authURL := oauthProvider.GetAuthURL(state, redirectURI)
	
	// TODO: Store state in cache/session for validation
	// For now, we'll return the state and expect it to be validated later
	
	return authURL, state, nil
}

func (s *AuthService) CompleteOAuth(provider, code, state, redirectURI string) (*AuthResponse, error) {
	oauthManager := NewOAuthManager(s.config)
	
	oauthProvider, err := oauthManager.GetProvider(provider)
	if err != nil {
		return nil, err
	}

	// TODO: Validate state from cache/session
	// For now, we'll skip state validation (not recommended for production)

	// Exchange code for token
	token, err := oauthProvider.ExchangeCode(code, redirectURI)
	if err != nil {
		return nil, err
	}

	// Get user info
	userInfo, err := oauthProvider.GetUserInfo(token)
	if err != nil {
		return nil, err
	}

	// Check if user exists
	var user models.User
	err = s.db.DB.Where("email = ?", userInfo.Email).First(&user).Error
	
	if err != nil {
		// User doesn't exist, create new user
		user = models.User{
			ID:        uuid.New(),
			Username:  s.generateUsername(userInfo.Username, userInfo.Name),
			Email:     userInfo.Email,
			FullName:  userInfo.Name,
			AvatarURL: userInfo.Avatar,
			IsActive:  true,
			IsAdmin:   false,
			EmailVerified: true, // OAuth emails are considered verified
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := s.db.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	} else {
		// Update existing user's info
		user.FullName = userInfo.Name
		user.AvatarURL = userInfo.Avatar
		now := time.Now()
		user.LastLoginAt = &now
		s.db.DB.Save(&user)
	}

	// Generate JWT token
	jwtToken, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""

	return &AuthResponse{
		User:  &user,
		Token: jwtToken,
	}, nil
}

func (s *AuthService) generateUsername(preferredUsername, fullName string) string {
	// Start with preferred username from OAuth provider
	if preferredUsername != "" {
		// Check if it's available
		var existingUser models.User
		if err := s.db.DB.Where("username = ?", preferredUsername).First(&existingUser).Error; err != nil {
			// Username is available
			return preferredUsername
		}
	}

	// Fall back to generating from full name
	base := strings.ToLower(strings.ReplaceAll(fullName, " ", ""))
	if base == "" {
		base = "user"
	}

	// Try the base username first
	var existingUser models.User
	if err := s.db.DB.Where("username = ?", base).First(&existingUser).Error; err != nil {
		return base
	}

	// If base is taken, append numbers until we find an available one
	for i := 1; i <= 1000; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if err := s.db.DB.Where("username = ?", candidate).First(&existingUser).Error; err != nil {
			return candidate
		}
	}

	// Last resort: use UUID
	return fmt.Sprintf("user_%s", uuid.New().String()[:8])
}