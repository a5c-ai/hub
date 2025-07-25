package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	ErrOIDCNotConfigured    = errors.New("OIDC not configured")
	ErrInvalidOIDCToken     = errors.New("invalid OIDC token")
	ErrOIDCDiscoveryFailed  = errors.New("OIDC discovery failed")
)

// OIDC Discovery Document
type OIDCDiscoveryDocument struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

// OIDC Claims
type OIDCClaims struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	Picture           string `json:"picture"`
	Groups            []string `json:"groups"`
	Roles             []string `json:"roles"`
}

// OIDC Token Response
type OIDCTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

type OIDCProvider struct {
	Name         string
	ClientID     string
	ClientSecret string
	IssuerURL    string
	RedirectURI  string
	Scopes       []string
	discovery    *OIDCDiscoveryDocument
}

type OIDCService struct {
	db         *gorm.DB
	config     *config.Config
	jwtManager *JWTManager
	authSvc    AuthService
	providers  map[string]*OIDCProvider
}

func NewOIDCService(db *gorm.DB, jwtManager *JWTManager, cfg *config.Config, authSvc AuthService) (*OIDCService, error) {
	service := &OIDCService{
		db:         db,
		config:     cfg,
		jwtManager: jwtManager,
		authSvc:    authSvc,
		providers:  make(map[string]*OIDCProvider),
	}

	// Initialize providers from configuration
	// This is a simplified setup - in practice, you'd configure multiple OIDC providers
	if cfg.SAML.Enabled { // Using SAML config as placeholder for OIDC config
		provider := &OIDCProvider{
			Name:         "generic-oidc",
			ClientID:     "your-client-id",     // This should come from config
			ClientSecret: "your-client-secret", // This should come from config
			IssuerURL:    "https://your-oidc-provider.com",
			RedirectURI:  "http://localhost:8080/auth/oidc/callback",
			Scopes:       []string{"openid", "profile", "email"},
		}

		// Discover endpoints
		if err := service.discoverProvider(provider); err != nil {
			return nil, fmt.Errorf("failed to discover OIDC provider: %w", err)
		}

		service.providers[provider.Name] = provider
	}

	return service, nil
}

func (s *OIDCService) discoverProvider(provider *OIDCProvider) error {
	discoveryURL := strings.TrimSuffix(provider.IssuerURL, "/") + "/.well-known/openid_configuration"
	
	resp, err := http.Get(discoveryURL)
	if err != nil {
		return fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrOIDCDiscoveryFailed
	}

	var discovery OIDCDiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return fmt.Errorf("failed to decode discovery document: %w", err)
	}

	provider.discovery = &discovery
	return nil
}

func (s *OIDCService) GetProvider(name string) (*OIDCProvider, error) {
	provider, exists := s.providers[name]
	if !exists {
		return nil, ErrOIDCNotConfigured
	}
	return provider, nil
}

func (s *OIDCService) GenerateAuthURL(providerName, state string) (string, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Add("client_id", provider.ClientID)
	params.Add("redirect_uri", provider.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", strings.Join(provider.Scopes, " "))
	params.Add("state", state)

	return provider.discovery.AuthorizationEndpoint + "?" + params.Encode(), nil
}

func (s *OIDCService) HandleCallback(ctx context.Context, providerName, code, state string) (*AuthResponse, error) {
	provider, err := s.GetProvider(providerName)
	if err != nil {
		return nil, err
	}

	// Exchange code for tokens
	tokenResponse, err := s.exchangeCode(provider, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user info from ID token or userinfo endpoint
	claims, err := s.getUserInfo(provider, tokenResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateOIDCUser(claims)
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

func (s *OIDCService) exchangeCode(provider *OIDCProvider, code string) (*OIDCTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", provider.ClientID)
	data.Set("client_secret", provider.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", provider.RedirectURI)

	req, err := http.NewRequest("POST", provider.discovery.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrOAuthCodeExchange
	}

	var tokenResponse OIDCTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func (s *OIDCService) getUserInfo(provider *OIDCProvider, tokenResponse *OIDCTokenResponse) (*OIDCClaims, error) {
	// For simplicity, we'll use the userinfo endpoint
	// In production, you should also validate and parse the ID token
	
	req, err := http.NewRequest("GET", provider.discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+tokenResponse.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidOIDCToken
	}

	var claims OIDCClaims
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (s *OIDCService) findOrCreateOIDCUser(claims *OIDCClaims) (*models.User, error) {
	// Try to find existing user by email
	var user models.User
	err := s.db.Where("email = ?", claims.Email).First(&user).Error

	if err == nil {
		// User exists, update profile if needed
		if user.FullName == "" && claims.Name != "" {
			user.FullName = claims.Name
		}
		if user.AvatarURL == "" && claims.Picture != "" {
			user.AvatarURL = claims.Picture
		}
		s.db.Save(&user)
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email[:strings.Index(claims.Email, "@")]
	}

	// Ensure username is unique
	username = s.ensureUniqueUsername(username)

	user = models.User{
		ID:            uuid.New(),
		Username:      username,
		Email:         claims.Email,
		FullName:      claims.Name,
		AvatarURL:     claims.Picture,
		EmailVerified: claims.EmailVerified,
		IsActive:      true,
		IsAdmin:       false,
		PasswordHash:  "", // OIDC users don't have passwords
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *OIDCService) ensureUniqueUsername(baseUsername string) string {
	username := baseUsername
	counter := 1

	for {
		var existingUser models.User
		if err := s.db.Where("username = ?", username).First(&existingUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			// Username is available
			return username
		}

		// Try with counter
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++

		if counter > 1000 {
			// Fallback to UUID
			return fmt.Sprintf("%s_%s", baseUsername, uuid.New().String()[:8])
		}
	}
}

// Just-in-Time (JIT) user provisioning
type JITProvisioningConfig struct {
	Enabled             bool              `json:"enabled"`
	DefaultRole         string            `json:"default_role"`
	AttributeMapping    map[string]string `json:"attribute_mapping"`
	GroupMapping        map[string]string `json:"group_mapping"`
	CreateOrganizations bool              `json:"create_organizations"`
}

func (s *OIDCService) ProvisionUser(claims *OIDCClaims, config *JITProvisioningConfig) (*models.User, error) {
	if !config.Enabled {
		return s.findOrCreateOIDCUser(claims)
	}

	// Enhanced user provisioning with group/role mapping
	user, err := s.findOrCreateOIDCUser(claims)
	if err != nil {
		return nil, err
	}

	// Map groups to roles if configured
	if len(config.GroupMapping) > 0 {
		for _, group := range claims.Groups {
			if role, exists := config.GroupMapping[group]; exists {
				// Apply role to user (this would need role management implementation)
				_ = role // TODO: implement role assignment
			}
		}
	}

	// Create organizations if configured
	if config.CreateOrganizations {
		for _, group := range claims.Groups {
			if err := s.createOrAssignOrganization(user.ID, group); err != nil {
				// Log error but don't fail the process
				fmt.Printf("Failed to create/assign organization %s for user %s: %v\n", group, user.Username, err)
			}
		}
	}

	return user, nil
}

// createOrAssignOrganization creates an organization if it doesn't exist and assigns the user
func (s *OIDCService) createOrAssignOrganization(userID uuid.UUID, groupName string) error {
	// Check if organization exists
	var org models.Organization
	err := s.db.Where("name = ?", groupName).First(&org).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new organization
		org = models.Organization{
			ID:          uuid.New(),
			Name:        groupName,
			DisplayName: groupName,
			Description: fmt.Sprintf("Organization created from OIDC group: %s", groupName),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		if err := s.db.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query organization: %w", err)
	}
	
	// Check if user is already a member
	var existingMember models.OrganizationMember
	err = s.db.Where("organization_id = ? AND user_id = ?", org.ID, userID).First(&existingMember).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Add user as organization member
		member := models.OrganizationMember{
			ID:             uuid.New(),
			OrganizationID: org.ID,
			UserID:         userID,
			Role:           models.OrgRoleMember, // Default role
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		
		if err := s.db.Create(&member).Error; err != nil {
			return fmt.Errorf("failed to add user to organization: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check organization membership: %w", err)
	}
	
	return nil
}