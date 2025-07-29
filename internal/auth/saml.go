package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSAMLNotConfigured = errors.New("SAML not configured")
	ErrInvalidSAMLResponse = errors.New("invalid SAML response")
	ErrSAMLSignatureInvalid = errors.New("SAML signature validation failed")
)

// SAML Response structures
type SAMLResponse struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol Response"`
	ID      string   `xml:"ID,attr"`
	Status  struct {
		StatusCode struct {
			Value string `xml:"Value,attr"`
		} `xml:"urn:oasis:names:tc:SAML:2.0:protocol StatusCode"`
	} `xml:"urn:oasis:names:tc:SAML:2.0:protocol Status"`
	Assertion struct {
		Subject struct {
			NameID struct {
				Value string `xml:",chardata"`
			} `xml:"urn:oasis:names:tc:SAML:2.0:assertion NameID"`
		} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Subject"`
		AttributeStatement struct {
			Attributes []struct {
				Name   string `xml:"Name,attr"`
				Values []struct {
					Value string `xml:",chardata"`
				} `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeValue"`
			} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Attribute"`
		} `xml:"urn:oasis:names:tc:SAML:2.0:assertion AttributeStatement"`
	} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Assertion"`
}

type SAMLRequest struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:protocol AuthnRequest"`
	ID      string   `xml:"ID,attr"`
	Version string   `xml:"Version,attr"`
	IssueInstant string `xml:"IssueInstant,attr"`
	Issuer  struct {
		Value string `xml:",chardata"`
	} `xml:"urn:oasis:names:tc:SAML:2.0:assertion Issuer"`
}

type SAMLUserInfo struct {
	ID       string
	Email    string
	Username string
	Name     string
	Groups   []string
}

type SAMLService struct {
	db         *gorm.DB
	config     *config.Config
	jwtManager *JWTManager
	authSvc    AuthService
	privateKey *rsa.PrivateKey
	cert       *x509.Certificate
}

func NewSAMLService(db *gorm.DB, jwtManager *JWTManager, cfg *config.Config, authSvc AuthService) (*SAMLService, error) {
	if !cfg.SAML.Enabled {
		return nil, ErrSAMLNotConfigured
	}

	service := &SAMLService{
		db:         db,
		config:     cfg,
		jwtManager: jwtManager,
		authSvc:    authSvc,
	}

	// Load private key and certificate if provided
	if cfg.SAML.PrivateKey != "" {
		if err := service.loadPrivateKey(cfg.SAML.PrivateKey); err != nil {
			return nil, fmt.Errorf("failed to load SAML private key: %w", err)
		}
	}

	if cfg.SAML.Certificate != "" {
		if err := service.loadCertificate(cfg.SAML.Certificate); err != nil {
			return nil, fmt.Errorf("failed to load SAML certificate: %w", err)
		}
	}

	return service, nil
}

func (s *SAMLService) loadPrivateKey(keyData string) error {
	block, _ := pem.Decode([]byte(keyData))
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return errors.New("key is not RSA private key")
		}
	}

	s.privateKey = key
	return nil
}

func (s *SAMLService) loadCertificate(certData string) error {
	block, _ := pem.Decode([]byte(certData))
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	s.cert = cert
	return nil
}

func (s *SAMLService) GenerateAuthRequest(relayState string) (string, error) {
	requestID := generateSAMLID()
	issueInstant := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	request := SAMLRequest{
		ID:           requestID,
		Version:      "2.0",
		IssueInstant: issueInstant,
	}
	request.Issuer.Value = s.config.SAML.EntityID

	// Marshal to XML
	xmlData, err := xml.MarshalIndent(request, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal SAML request: %w", err)
	}

	// Base64 encode
	encodedRequest := base64.StdEncoding.EncodeToString(xmlData)

	// Build redirect URL
	params := url.Values{}
	params.Add("SAMLRequest", encodedRequest)
	if relayState != "" {
		params.Add("RelayState", relayState)
	}

	redirectURL := s.config.SAML.SSOURL + "?" + params.Encode()
	return redirectURL, nil
}

func (s *SAMLService) ProcessResponse(samlResponse string, relayState string) (*AuthResponse, error) {
	// Decode base64
	decodedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode SAML response: %w", err)
	}

	// Parse XML
	var response SAMLResponse
	if err := xml.Unmarshal(decodedResponse, &response); err != nil {
		return nil, fmt.Errorf("failed to parse SAML response: %w", err)
	}

	// Check status
	if response.Status.StatusCode.Value != "urn:oasis:names:tc:SAML:2.0:status:Success" {
		return nil, ErrInvalidSAMLResponse
	}

	// Extract user info
	userInfo := s.extractUserInfo(&response)

	// Find or create user
	user, err := s.findOrCreateSAMLUser(userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create a session to generate a secure refresh token
	session, err := s.authSvc.(*authService).sessionService.CreateSession(user.ID, "", "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to create session for refresh token: %w", err)
	}
	refreshToken := session.RefreshToken

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

func (s *SAMLService) extractUserInfo(response *SAMLResponse) *SAMLUserInfo {
	userInfo := &SAMLUserInfo{
		ID: response.Assertion.Subject.NameID.Value,
	}

	// Map attributes based on configuration
	for _, attr := range response.Assertion.AttributeStatement.Attributes {
		if len(attr.Values) == 0 {
			continue
		}

		// Check attribute mapping
		if mappedField, exists := s.config.SAML.AttributeMap[attr.Name]; exists {
			switch mappedField {
			case "email":
				userInfo.Email = attr.Values[0].Value
			case "username":
				userInfo.Username = attr.Values[0].Value
			case "name":
				userInfo.Name = attr.Values[0].Value
			case "groups":
				for _, val := range attr.Values {
					userInfo.Groups = append(userInfo.Groups, val.Value)
				}
			}
		}

		// Default mappings
		switch attr.Name {
		case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
			if userInfo.Email == "" {
				userInfo.Email = attr.Values[0].Value
			}
		case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name":
			if userInfo.Username == "" {
				userInfo.Username = attr.Values[0].Value
			}
		case "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname":
			if userInfo.Name == "" {
				userInfo.Name = attr.Values[0].Value
			}
		}
	}

	return userInfo
}

func (s *SAMLService) findOrCreateSAMLUser(userInfo *SAMLUserInfo) (*models.User, error) {
	// Try to find existing user by email
	var user models.User
	err := s.db.Where("email = ?", userInfo.Email).First(&user).Error

	if err == nil {
		// User exists, update profile if needed
		if user.FullName == "" && userInfo.Name != "" {
			user.FullName = userInfo.Name
		}
		s.db.Save(&user)
		
		// Handle group-based organization assignment for existing users
		if len(userInfo.Groups) > 0 {
			for _, group := range userInfo.Groups {
				if err := s.createOrAssignOrganization(user.ID, group); err != nil {
					// Log error but don't fail the process
					fmt.Printf("Failed to create/assign organization %s for user %s: %v\n", group, user.Username, err)
				}
			}
		}
		
		return &user, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	username := userInfo.Username
	if username == "" {
		username = userInfo.Email[:strings.Index(userInfo.Email, "@")]
	}

	// Ensure username is unique
	username = s.ensureUniqueUsername(username)

	user = models.User{
		ID:            uuid.New(),
		Username:      username,
		Email:         userInfo.Email,
		FullName:      userInfo.Name,
		EmailVerified: true, // SAML emails are considered verified
		IsActive:      true,
		IsAdmin:       false,
		PasswordHash:  "", // SAML users don't have passwords
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	// Handle group-based organization assignment for new users
	if len(userInfo.Groups) > 0 {
		for _, group := range userInfo.Groups {
			if err := s.createOrAssignOrganization(user.ID, group); err != nil {
				// Log error but don't fail the process
				fmt.Printf("Failed to create/assign organization %s for user %s: %v\n", group, user.Username, err)
			}
		}
	}

	return &user, nil
}

func (s *SAMLService) ensureUniqueUsername(baseUsername string) string {
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

func generateSAMLID() string {
	return fmt.Sprintf("_%s", uuid.New().String())
}

// SAML metadata generation
func (s *SAMLService) GenerateMetadata() (string, error) {
	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata"
                     entityID="%s">
  <md:SPSSODescriptor
      AuthnRequestsSigned="false"
      WantAssertionsSigned="true"
      protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:AssertionConsumerService
        Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
        Location="%s/auth/saml/acs"
        index="0"
        isDefault="true"/>
  </md:SPSSODescriptor>
</md:EntityDescriptor>`, s.config.SAML.EntityID, getSAMLBaseURL())

	return metadata, nil
}

// Helper function that should be implemented based on your base URL configuration
func getSAMLBaseURL() string {
	// This should be configurable
	return "http://localhost:8080"
}

// createOrAssignOrganization creates an organization if it doesn't exist and assigns the user
func (s *SAMLService) createOrAssignOrganization(userID uuid.UUID, groupName string) error {
	// Check if organization exists
	var org models.Organization
	err := s.db.Where("name = ?", groupName).First(&org).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new organization
		org = models.Organization{
			ID:          uuid.New(),
			Name:        groupName,
			DisplayName: groupName,
			Description: fmt.Sprintf("Organization created from SAML group: %s", groupName),
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
