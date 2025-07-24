package auth

import (
	"context"
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
	"gorm.io/gorm"
)

type OAuthConfig struct {
	GitHub    GitHubConfig    `mapstructure:"github"`
	Google    GoogleConfig    `mapstructure:"google"`
	Microsoft MicrosoftConfig `mapstructure:"microsoft"`
}

type GitHubConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type GoogleConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type MicrosoftConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
	TenantID     string `mapstructure:"tenant_id"`
}

type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type OAuthUserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(code string) (*OAuthToken, error)
	GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error)
	GetProviderName() string
}

type GitHubProvider struct {
	config GitHubConfig
}

type GoogleProvider struct {
	config GoogleConfig
}

type MicrosoftProvider struct {
	config MicrosoftConfig
}

type OAuthService struct {
	db        *gorm.DB
	providers map[string]OAuthProvider
	authSvc   AuthService
}

func NewOAuthService(db *gorm.DB, oauthConfig OAuthConfig, authSvc AuthService) *OAuthService {
	providers := make(map[string]OAuthProvider)
	
	if oauthConfig.GitHub.ClientID != "" {
		providers["github"] = &GitHubProvider{config: oauthConfig.GitHub}
	}
	
	if oauthConfig.Google.ClientID != "" {
		providers["google"] = &GoogleProvider{config: oauthConfig.Google}
	}
	
	if oauthConfig.Microsoft.ClientID != "" {
		providers["microsoft"] = &MicrosoftProvider{config: oauthConfig.Microsoft}
	}

	return &OAuthService{
		db:        db,
		providers: providers,
		authSvc:   authSvc,
	}
}

func (s *OAuthService) GetProvider(name string) (OAuthProvider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("OAuth provider %s not configured", name)
	}
	return provider, nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code, state string) (*AuthResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Exchange code for token
	token, err := provider.ExchangeCode(code)
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
	jwtManager := NewJWTManager(config.JWT{Secret: "temp", ExpirationHour: 24}) // TODO: use proper config
	accessToken, err := jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(user)

	return &AuthResponse{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   24 * 3600, // 24 hours
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
	username := userInfo.Username
	if username == "" {
		username = strings.Split(userInfo.Email, "@")[0]
	}

	// Ensure username is unique
	for i := 0; i < 100; i++ {
		var existing models.User
		testUsername := username
		if i > 0 {
			testUsername = fmt.Sprintf("%s%d", username, i)
		}
		
		err := s.db.Where("username = ?", testUsername).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			username = testUsername
			break
		}
	}

	user = models.User{
		Username:      username,
		Email:         userInfo.Email,
		FullName:      userInfo.Name,
		AvatarURL:     userInfo.Avatar,
		EmailVerified: true, // OAuth emails are considered verified
		IsActive:      true,
		IsAdmin:       false,
		PasswordHash:  "", // OAuth users don't have passwords
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GitHub Provider Implementation
func (p *GitHubProvider) GetProviderName() string {
	return "github"
}

func (p *GitHubProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", p.config.ClientID)
	params.Add("redirect_uri", p.config.RedirectURL)
	params.Add("scope", "user:email")
	params.Add("state", state)
	
	return "https://github.com/login/oauth/authorize?" + params.Encode()
}

func (p *GitHubProvider) ExchangeCode(code string) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("code", code)

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

	// Get primary email if not provided
	if githubUser.Email == "" {
		email, err := p.getPrimaryEmail(token.AccessToken)
		if err == nil {
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

	return "", errors.New("no email found")
}

// Google Provider Implementation (simplified)
func (p *GoogleProvider) GetProviderName() string {
	return "google"
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", p.config.ClientID)
	params.Add("redirect_uri", p.config.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	
	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (p *GoogleProvider) ExchangeCode(code string) (*OAuthToken, error) {
	// TODO: Implement Google OAuth token exchange
	return nil, errors.New("Google OAuth not fully implemented")
}

func (p *GoogleProvider) GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	// TODO: Implement Google user info retrieval
	return nil, errors.New("Google OAuth not fully implemented")
}

// Microsoft Provider Implementation (simplified)
func (p *MicrosoftProvider) GetProviderName() string {
	return "microsoft"
}

func (p *MicrosoftProvider) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", p.config.ClientID)
	params.Add("redirect_uri", p.config.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "openid email profile")
	params.Add("state", state)
	
	baseURL := "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
	if p.config.TenantID != "" {
		baseURL = fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", p.config.TenantID)
	}
	
	return baseURL + "?" + params.Encode()
}

func (p *MicrosoftProvider) ExchangeCode(code string) (*OAuthToken, error) {
	// TODO: Implement Microsoft OAuth token exchange
	return nil, errors.New("Microsoft OAuth not fully implemented")
}

func (p *MicrosoftProvider) GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error) {
	// TODO: Implement Microsoft user info retrieval
	return nil, errors.New("Microsoft OAuth not fully implemented")
}