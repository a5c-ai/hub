package auth

import (
	"context"
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
	"gorm.io/gorm"
)

var (
	ErrOAuthProviderNotConfigured = errors.New("oauth provider not configured")
	ErrInvalidOAuthState          = errors.New("invalid oauth state")
	ErrOAuthCodeExchange          = errors.New("oauth code exchange failed")
	ErrOAuthUserInfo              = errors.New("failed to get oauth user info")
)

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

type OAuthUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

type OAuthProvider interface {
	GetAuthURL(state string, redirectURI string) string
	ExchangeCode(code, redirectURI string) (*OAuthToken, error)
	GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error)
	GetProviderName() string
}

type OAuthService struct {
	db         *gorm.DB
	providers  map[string]OAuthProvider
	authSvc    AuthService
	jwtManager *JWTManager
	config     *config.Config
}

func NewOAuthService(db *gorm.DB, jwtManager *JWTManager, cfg *config.Config, authSvc AuthService) *OAuthService {
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

	// Initialize Microsoft provider if configured
	if cfg.OAuth.Microsoft.ClientID != "" && cfg.OAuth.Microsoft.ClientSecret != "" {
		providers["microsoft"] = &MicrosoftProvider{
			ClientID:     cfg.OAuth.Microsoft.ClientID,
			ClientSecret: cfg.OAuth.Microsoft.ClientSecret,
			TenantID:     cfg.OAuth.Microsoft.TenantID,
		}
	}

	return &OAuthService{
		db:         db,
		providers:  providers,
		authSvc:    authSvc,
		jwtManager: jwtManager,
		config:     cfg,
	}
}

func (s *OAuthService) GetProvider(name string) (OAuthProvider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, ErrOAuthProviderNotConfigured
	}
	return provider, nil
}

func (s *OAuthService) GenerateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *OAuthService) InitiateOAuth(provider, redirectURI string) (string, string, error) {
	oauthProvider, err := s.GetProvider(provider)
	if err != nil {
		return "", "", err
	}

	state, err := s.GenerateState()
	if err != nil {
		return "", "", err
	}

	authURL := oauthProvider.GetAuthURL(state, redirectURI)
	
	// TODO: Store state in cache/session for validation
	// For now, we'll return the state and expect it to be validated later
	
	return authURL, state, nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code, state, redirectURI string) (*AuthResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// TODO: Validate state from cache/session
	// For now, we'll skip state validation (not recommended for production)

	// Exchange code for token
	token, err := provider.ExchangeCode(code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info
	userInfo, err := provider.GetUserInfo(token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateOAuthUser(userInfo, providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate our own tokens
	accessToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(user) // TODO: implement proper refresh token logic
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(user)

	// Remove sensitive information before returning
	user.PasswordHash = ""

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Duration(s.config.JWT.ExpirationHour) * time.Hour / time.Second),
	}, nil
}

func (s *OAuthService) findOrCreateOAuthUser(userInfo *OAuthUserInfo, providerName string) (*models.User, error) {
	// Try to find existing user by email
	var user models.User
	err := s.db.Where("email = ?", userInfo.Email).First(&user).Error
	
	if err == nil {
		// User exists, update profile if needed
		if user.FullName == "" && userInfo.Name != "" {
			user.FullName = userInfo.Name
		}
		if user.AvatarURL == "" && userInfo.Avatar != "" {
			user.AvatarURL = userInfo.Avatar
		}
		s.db.Save(&user)
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	username := s.generateUsername(userInfo.Username, userInfo.Name)

	user = models.User{
		ID:            uuid.New(),
		Username:      username,
		Email:         userInfo.Email,
		FullName:      userInfo.Name,
		AvatarURL:     userInfo.Avatar,
		EmailVerified: true, // OAuth emails are considered verified
		IsActive:      true,
		IsAdmin:       false,
		PasswordHash:  "", // OAuth users don't have passwords
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *OAuthService) generateUsername(preferredUsername, fullName string) string {
	// Start with preferred username from OAuth provider
	if preferredUsername != "" {
		// Check if it's available
		var existingUser models.User
		if err := s.db.Where("username = ?", preferredUsername).First(&existingUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
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
	if err := s.db.Where("username = ?", base).First(&existingUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return base
	}

	// If base is taken, append numbers until we find an available one
	for i := 1; i <= 1000; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if err := s.db.Where("username = ?", candidate).First(&existingUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			return candidate
		}
	}

	// Last resort: use UUID
	return fmt.Sprintf("user_%s", uuid.New().String()[:8])
}

// GitHub OAuth Provider
type GitHubProvider struct {
	ClientID     string
	ClientSecret string
}

func (p *GitHubProvider) GetProviderName() string {
	return "github"
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token OAuthToken
	if err := json.Unmarshal(body, &token); err != nil {
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

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, err
	}

	// GitHub doesn't always return public email, fetch it separately
	if githubUser.Email == "" {
		if email, err := p.getPrimaryEmail(token.AccessToken); err == nil {
			githubUser.Email = email
		}
	}

	return &OAuthUserInfo{
		ID:       fmt.Sprintf("%d", githubUser.ID),
		Username: githubUser.Login,
		Email:    githubUser.Email,
		Name:     githubUser.Name,
		Avatar:   githubUser.AvatarURL,
	}, nil
}

func (p *GitHubProvider) getPrimaryEmail(accessToken string) (string, error) {
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
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

func (p *GoogleProvider) GetProviderName() string {
	return "google"
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

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:       googleUser.ID,
		Username: strings.Split(googleUser.Email, "@")[0], // Generate username from email
		Email:    googleUser.Email,
		Name:     googleUser.Name,
		Avatar:   googleUser.Picture,
	}, nil
}

// Microsoft OAuth Provider
type MicrosoftProvider struct {
	ClientID     string
	ClientSecret string
	TenantID     string
}

func (p *MicrosoftProvider) GetProviderName() string {
	return "microsoft"
}

func (p *MicrosoftProvider) GetAuthURL(state, redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", p.ClientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	
	baseURL := "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	if p.TenantID != "" {
		baseURL = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", p.TenantID)
	}
	
	return baseURL + "?" + params.Encode()
}

func (p *MicrosoftProvider) ExchangeCode(code, redirectURI string) (*OAuthToken, error) {
	// TODO: Implement Microsoft OAuth token exchange
	return nil, errors.New("Microsoft OAuth not fully implemented")
}

func (p *MicrosoftProvider) GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	// TODO: Implement Microsoft user info retrieval
	return nil, errors.New("Microsoft OAuth not fully implemented")
}