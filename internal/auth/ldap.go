package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrLDAPNotConfigured = errors.New("LDAP not configured")
	ErrLDAPConnection    = errors.New("LDAP connection failed")
	ErrLDAPAuth          = errors.New("LDAP authentication failed")
	ErrLDAPUserNotFound  = errors.New("LDAP user not found")
)

// LDAP User Information
type LDAPUserInfo struct {
	DN          string
	Username    string
	Email       string
	FirstName   string
	LastName    string
	DisplayName string
	Groups      []string
	Attributes  map[string][]string
}

// LDAP Connection interface for testing
type LDAPConnection interface {
	Bind(username, password string) error
	Search(baseDN, filter string, attributes []string) ([]LDAPSearchResult, error)
	Close() error
}

type LDAPSearchResult struct {
	DN         string
	Attributes map[string][]string
}

// Mock LDAP connection for development/testing
type MockLDAPConnection struct {
	users map[string]MockLDAPUser
}

type MockLDAPUser struct {
	DN          string
	Password    string
	Email       string
	FirstName   string
	LastName    string
	DisplayName string
	Groups      []string
}

func NewMockLDAPConnection() *MockLDAPConnection {
	return &MockLDAPConnection{
		users: map[string]MockLDAPUser{
			"testuser": {
				DN:          "cn=testuser,ou=users,dc=example,dc=com",
				Password:    "testpass",
				Email:       "testuser@example.com",
				FirstName:   "Test",
				LastName:    "User",
				DisplayName: "Test User",
				Groups:      []string{"developers", "users"},
			},
			"admin": {
				DN:          "cn=admin,ou=users,dc=example,dc=com",
				Password:    "adminpass",
				Email:       "admin@example.com",
				FirstName:   "Admin",
				LastName:    "User",
				DisplayName: "Admin User",
				Groups:      []string{"admins", "users"},
			},
		},
	}
}

func (m *MockLDAPConnection) Bind(username, password string) error {
	user, exists := m.users[username]
	if !exists {
		return ErrLDAPUserNotFound
	}
	if user.Password != password {
		return ErrLDAPAuth
	}
	return nil
}

func (m *MockLDAPConnection) Search(baseDN, filter string, attributes []string) ([]LDAPSearchResult, error) {
	var results []LDAPSearchResult
	
	for username, user := range m.users {
		// Simple filter matching
		if strings.Contains(filter, username) || strings.Contains(filter, user.Email) {
			attrs := map[string][]string{
				"cn":          {username},
				"mail":        {user.Email},
				"givenName":   {user.FirstName},
				"sn":          {user.LastName},
				"displayName": {user.DisplayName},
				"memberOf":    user.Groups,
			}
			
			results = append(results, LDAPSearchResult{
				DN:         user.DN,
				Attributes: attrs,
			})
		}
	}
	
	return results, nil
}

func (m *MockLDAPConnection) Close() error {
	return nil
}

type LDAPService struct {
	db         *gorm.DB
	config     *config.Config
	jwtManager *JWTManager
	authSvc    AuthService
}

func NewLDAPService(db *gorm.DB, jwtManager *JWTManager, cfg *config.Config, authSvc AuthService) (*LDAPService, error) {
	if !cfg.LDAP.Enabled {
		return nil, ErrLDAPNotConfigured
	}

	return &LDAPService{
		db:         db,
		config:     cfg,
		jwtManager: jwtManager,
		authSvc:    authSvc,
	}, nil
}

func (s *LDAPService) createConnection() (LDAPConnection, error) {
	// For development, use mock connection
	if s.config.Environment == "development" || s.config.LDAP.Host == "" {
		return NewMockLDAPConnection(), nil
	}

	// In production, you would use a real LDAP library like go-ldap
	// Here's a placeholder for the real implementation:
	
	// conn, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%d", s.config.LDAP.Host, s.config.LDAP.Port))
	// if err != nil {
	//     return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	// }
	
	// Bind with service account if configured
	// if s.config.LDAP.BindDN != "" && s.config.LDAP.BindPassword != "" {
	//     err = conn.Bind(s.config.LDAP.BindDN, s.config.LDAP.BindPassword)
	//     if err != nil {
	//         conn.Close()
	//         return nil, fmt.Errorf("failed to bind to LDAP: %w", err)
	//     }
	// }
	
	// return conn, nil
	
	return NewMockLDAPConnection(), nil
}

func (s *LDAPService) Authenticate(username, password string) (*AuthResponse, error) {
	conn, err := s.createConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Search for user
	userInfo, err := s.searchUser(conn, username)
	if err != nil {
		return nil, err
	}

	// Authenticate user
	userConn, err := s.createConnection()
	if err != nil {
		return nil, err
	}
	defer userConn.Close()

	if err := userConn.Bind(userInfo.DN, password); err != nil {
		return nil, ErrLDAPAuth
	}

	// Find or create user in local database
	user, err := s.findOrCreateLDAPUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwtManager.GenerateToken(user)
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

func (s *LDAPService) searchUser(conn LDAPConnection, username string) (*LDAPUserInfo, error) {
	// Build search filter
	filter := s.config.LDAP.UserFilter
	if filter == "" {
		filter = "(uid=%s)"
	}
	filter = fmt.Sprintf(filter, username)

	// Search attributes
	attributes := []string{"cn", "mail", "givenName", "sn", "displayName", "memberOf"}
	
	results, err := conn.Search(s.config.LDAP.BaseDN, filter, attributes)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrLDAPUserNotFound
	}

	if len(results) > 1 {
		return nil, fmt.Errorf("multiple users found for username: %s", username)
	}

	result := results[0]
	userInfo := &LDAPUserInfo{
		DN:         result.DN,
		Username:   username,
		Attributes: result.Attributes,
	}

	// Map attributes based on configuration
	if s.config.LDAP.AttributeMap != nil {
		if emailAttr, exists := s.config.LDAP.AttributeMap["email"]; exists {
			if values, found := result.Attributes[emailAttr]; found && len(values) > 0 {
				userInfo.Email = values[0]
			}
		}
		if fnameAttr, exists := s.config.LDAP.AttributeMap["first_name"]; exists {
			if values, found := result.Attributes[fnameAttr]; found && len(values) > 0 {
				userInfo.FirstName = values[0]
			}
		}
		if lnameAttr, exists := s.config.LDAP.AttributeMap["last_name"]; exists {
			if values, found := result.Attributes[lnameAttr]; found && len(values) > 0 {
				userInfo.LastName = values[0]
			}
		}
		if dnameAttr, exists := s.config.LDAP.AttributeMap["display_name"]; exists {
			if values, found := result.Attributes[dnameAttr]; found && len(values) > 0 {
				userInfo.DisplayName = values[0]
			}
		}
	} else {
		// Default attribute mapping
		if values, found := result.Attributes["mail"]; found && len(values) > 0 {
			userInfo.Email = values[0]
		}
		if values, found := result.Attributes["givenName"]; found && len(values) > 0 {
			userInfo.FirstName = values[0]
		}
		if values, found := result.Attributes["sn"]; found && len(values) > 0 {
			userInfo.LastName = values[0]
		}
		if values, found := result.Attributes["displayName"]; found && len(values) > 0 {
			userInfo.DisplayName = values[0]
		}
	}

	// Extract groups
	if values, found := result.Attributes["memberOf"]; found {
		userInfo.Groups = values
	}

	return userInfo, nil
}

func (s *LDAPService) findOrCreateLDAPUser(userInfo *LDAPUserInfo) (*models.User, error) {
	// Try to find existing user by email or username
	var user models.User
	err := s.db.Where("email = ? OR username = ?", userInfo.Email, userInfo.Username).First(&user).Error

	if err == nil {
		// User exists, update profile
		user.FullName = s.buildFullName(userInfo.FirstName, userInfo.LastName, userInfo.DisplayName)
		s.db.Save(&user)
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	fullName := s.buildFullName(userInfo.FirstName, userInfo.LastName, userInfo.DisplayName)
	
	user = models.User{
		ID:            uuid.New(),
		Username:      userInfo.Username,
		Email:         userInfo.Email,
		FullName:      fullName,
		EmailVerified: true, // LDAP users are considered verified
		IsActive:      true,
		IsAdmin:       s.isAdminUser(userInfo),
		PasswordHash:  "", // LDAP users don't have local passwords
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *LDAPService) buildFullName(firstName, lastName, displayName string) string {
	if displayName != "" {
		return displayName
	}
	if firstName != "" && lastName != "" {
		return fmt.Sprintf("%s %s", firstName, lastName)
	}
	if firstName != "" {
		return firstName
	}
	if lastName != "" {
		return lastName
	}
	return ""
}

func (s *LDAPService) isAdminUser(userInfo *LDAPUserInfo) bool {
	// Check if user is in admin groups
	adminGroups := []string{"admins", "administrators", "domain admins"}
	
	for _, group := range userInfo.Groups {
		groupName := strings.ToLower(s.extractGroupName(group))
		for _, adminGroup := range adminGroups {
			if groupName == adminGroup {
				return true
			}
		}
	}
	
	return false
}

func (s *LDAPService) extractGroupName(groupDN string) string {
	// Extract group name from DN (e.g., "cn=admins,ou=groups,dc=example,dc=com" -> "admins")
	parts := strings.Split(groupDN, ",")
	if len(parts) > 0 {
		cnPart := parts[0]
		if strings.HasPrefix(strings.ToLower(cnPart), "cn=") {
			return cnPart[3:]
		}
	}
	return groupDN
}

// LDAP user synchronization
func (s *LDAPService) SyncUsers() error {
	conn, err := s.createConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Search for all users
	filter := "(objectClass=person)"
	attributes := []string{"cn", "mail", "givenName", "sn", "displayName", "memberOf"}
	
	results, err := conn.Search(s.config.LDAP.BaseDN, filter, attributes)
	if err != nil {
		return fmt.Errorf("LDAP sync search failed: %w", err)
	}

	for _, result := range results {
		// Extract username from CN or other attribute
		username := ""
		if values, found := result.Attributes["cn"]; found && len(values) > 0 {
			username = values[0]
		}
		
		if username == "" {
			continue
		}

		userInfo := &LDAPUserInfo{
			DN:         result.DN,
			Username:   username,
			Attributes: result.Attributes,
		}

		// Map attributes (similar to searchUser)
		if values, found := result.Attributes["mail"]; found && len(values) > 0 {
			userInfo.Email = values[0]
		}
		if values, found := result.Attributes["givenName"]; found && len(values) > 0 {
			userInfo.FirstName = values[0]
		}
		if values, found := result.Attributes["sn"]; found && len(values) > 0 {
			userInfo.LastName = values[0]
		}
		if values, found := result.Attributes["displayName"]; found && len(values) > 0 {
			userInfo.DisplayName = values[0]
		}
		if values, found := result.Attributes["memberOf"]; found {
			userInfo.Groups = values
		}

		// Update or create user
		_, err := s.findOrCreateLDAPUser(userInfo)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to sync user %s: %v\n", username, err)
		}
	}

	return nil
}

// Test LDAP connection
func (s *LDAPService) TestConnection() error {
	conn, err := s.createConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Try to search for a single entry to test the connection
	results, err := conn.Search(s.config.LDAP.BaseDN, "(objectClass=*)", []string{"cn"})
	if err != nil {
		return fmt.Errorf("LDAP connection test failed: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("LDAP connection test returned no results")
	}

	return nil
}