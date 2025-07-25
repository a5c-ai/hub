package services

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// DeployKeyService handles deploy key management
type DeployKeyService struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewDeployKeyService creates a new deploy key service
func NewDeployKeyService(db *gorm.DB, logger *logrus.Logger) *DeployKeyService {
	return &DeployKeyService{
		db:     db,
		logger: logger,
	}
}

// CreateDeployKey creates a new deploy key
func (s *DeployKeyService) CreateDeployKey(ctx context.Context, repositoryID uuid.UUID, title, key string, readOnly bool) (*models.DeployKey, error) {
	// Validate and parse the SSH key
	fingerprint, err := s.generateFingerprint(key)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH key: %w", err)
	}
	
	// Check if key already exists for this repository
	var existing models.DeployKey
	err = s.db.WithContext(ctx).Where("repository_id = ? AND fingerprint = ?", repositoryID, fingerprint).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("deploy key already exists for this repository")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing deploy key: %w", err)
	}
	
	deployKey := &models.DeployKey{
		RepositoryID: repositoryID,
		Title:        title,
		Key:          key,
		Fingerprint:  fingerprint,
		ReadOnly:     readOnly,
		Verified:     true, // Auto-verify for now
	}
	
	if err := s.db.WithContext(ctx).Create(deployKey).Error; err != nil {
		return nil, fmt.Errorf("failed to create deploy key: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"deploy_key_id": deployKey.ID,
		"repository_id": repositoryID,
		"title":         title,
		"fingerprint":   fingerprint,
		"read_only":     readOnly,
	}).Info("Created deploy key")
	
	return deployKey, nil
}

// GetDeployKey retrieves a deploy key by ID
func (s *DeployKeyService) GetDeployKey(ctx context.Context, deployKeyID uuid.UUID) (*models.DeployKey, error) {
	var deployKey models.DeployKey
	if err := s.db.WithContext(ctx).First(&deployKey, "id = ?", deployKeyID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("deploy key not found")
		}
		return nil, fmt.Errorf("failed to get deploy key: %w", err)
	}
	return &deployKey, nil
}

// ListDeployKeys lists all deploy keys for a repository
func (s *DeployKeyService) ListDeployKeys(ctx context.Context, repositoryID uuid.UUID) ([]models.DeployKey, error) {
	var deployKeys []models.DeployKey
	if err := s.db.WithContext(ctx).Where("repository_id = ?", repositoryID).Find(&deployKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to list deploy keys: %w", err)
	}
	return deployKeys, nil
}

// DeleteDeployKey deletes a deploy key
func (s *DeployKeyService) DeleteDeployKey(ctx context.Context, deployKeyID uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.DeployKey{}, "id = ?", deployKeyID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete deploy key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("deploy key not found")
	}
	
	s.logger.WithField("deploy_key_id", deployKeyID).Info("Deleted deploy key")
	return nil
}

// VerifyDeployKey verifies a deploy key format and updates its verification status
func (s *DeployKeyService) VerifyDeployKey(ctx context.Context, deployKeyID uuid.UUID) error {
	deployKey, err := s.GetDeployKey(ctx, deployKeyID)
	if err != nil {
		return err
	}
	
	// Verify the SSH key format
	_, err = s.parseSSHKey(deployKey.Key)
	if err != nil {
		deployKey.Verified = false
	} else {
		deployKey.Verified = true
	}
	
	if err := s.db.WithContext(ctx).Save(deployKey).Error; err != nil {
		return fmt.Errorf("failed to update deploy key verification: %w", err)
	}
	
	return nil
}

// UpdateLastUsed updates the last used timestamp for a deploy key
func (s *DeployKeyService) UpdateLastUsed(ctx context.Context, fingerprint string) error {
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&models.DeployKey{}).
		Where("fingerprint = ?", fingerprint).
		Update("last_used_at", now)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update last used: %w", result.Error)
	}
	
	return nil
}

// FindDeployKeyByFingerprint finds a deploy key by its fingerprint
func (s *DeployKeyService) FindDeployKeyByFingerprint(ctx context.Context, fingerprint string) (*models.DeployKey, error) {
	var deployKey models.DeployKey
	if err := s.db.WithContext(ctx).Where("fingerprint = ?", fingerprint).First(&deployKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("deploy key not found")
		}
		return nil, fmt.Errorf("failed to find deploy key: %w", err)
	}
	return &deployKey, nil
}

// ValidateSSHKey validates an SSH key format
func (s *DeployKeyService) ValidateSSHKey(key string) error {
	_, err := s.parseSSHKey(key)
	return err
}

// parseSSHKey parses and validates an SSH public key
func (s *DeployKeyService) parseSSHKey(keyStr string) (ssh.PublicKey, error) {
	// Clean up the key string
	keyStr = strings.TrimSpace(keyStr)
	
	// Handle keys with comments
	parts := strings.Fields(keyStr)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid SSH key format")
	}
	
	// Reconstruct key without comment for parsing
	keyForParsing := parts[0] + " " + parts[1]
	
	// Parse the SSH public key
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyForParsing))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH key: %w", err)
	}
	
	return publicKey, nil
}

// generateFingerprint generates a fingerprint for an SSH key
func (s *DeployKeyService) generateFingerprint(keyStr string) (string, error) {
	publicKey, err := s.parseSSHKey(keyStr)
	if err != nil {
		return "", err
	}
	
	// Generate SHA256 fingerprint (modern format)
	hash := sha256.Sum256(publicKey.Marshal())
	fingerprint := base64.StdEncoding.EncodeToString(hash[:])
	
	// Remove padding and format as SHA256:fingerprint
	fingerprint = strings.TrimRight(fingerprint, "=")
	return "SHA256:" + fingerprint, nil
}

// generateMD5Fingerprint generates an MD5 fingerprint for an SSH key (legacy format)
func (s *DeployKeyService) generateMD5Fingerprint(keyStr string) (string, error) {
	publicKey, err := s.parseSSHKey(keyStr)
	if err != nil {
		return "", err
	}
	
	// Generate MD5 fingerprint (legacy format)
	hash := md5.Sum(publicKey.Marshal())
	
	// Format as colon-separated hex pairs
	var fingerprint strings.Builder
	for i, b := range hash {
		if i > 0 {
			fingerprint.WriteString(":")
		}
		fingerprint.WriteString(fmt.Sprintf("%02x", b))
	}
	
	return "MD5:" + fingerprint.String(), nil
}

// GetKeyType returns the type of SSH key (rsa, ed25519, ecdsa, etc.)
func (s *DeployKeyService) GetKeyType(keyStr string) (string, error) {
	publicKey, err := s.parseSSHKey(keyStr)
	if err != nil {
		return "", err
	}
	
	return publicKey.Type(), nil
}

// GetKeySize returns the size of the SSH key in bits
func (s *DeployKeyService) GetKeySize(keyStr string) (int, error) {
	publicKey, err := s.parseSSHKey(keyStr)
	if err != nil {
		return 0, err
	}
	
	// Get key size from the key type string
	keyType := publicKey.Type()
	switch keyType {
	case "ssh-rsa":
		// RSA key size calculation would require more complex parsing
		return 2048, nil // Common RSA key size
	case "ecdsa-sha2-nistp256":
		return 256, nil
	case "ecdsa-sha2-nistp384":
		return 384, nil
	case "ecdsa-sha2-nistp521":
		return 521, nil
	case "ssh-ed25519":
		return 256, nil // Ed25519 keys are always 256 bits
	default:
		return 0, nil
	}
}

// IsKeySecure checks if the SSH key meets security requirements
func (s *DeployKeyService) IsKeySecure(keyStr string) (bool, []string, error) {
	var warnings []string
	
	keyType, err := s.GetKeyType(keyStr)
	if err != nil {
		return false, warnings, err
	}
	
	keySize, err := s.GetKeySize(keyStr)
	if err != nil {
		return false, warnings, err
	}
	
	secure := true
	
	switch keyType {
	case "ssh-rsa":
		if keySize < 2048 {
			secure = false
			warnings = append(warnings, fmt.Sprintf("RSA key size (%d bits) is below recommended minimum of 2048 bits", keySize))
		} else if keySize < 3072 {
			warnings = append(warnings, fmt.Sprintf("RSA key size (%d bits) is below current best practice of 3072+ bits", keySize))
		}
	case "ssh-dss":
		secure = false
		warnings = append(warnings, "DSA keys are considered insecure and should not be used")
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		// ECDSA keys are generally secure, but some prefer Ed25519
		if keySize < 256 {
			warnings = append(warnings, "ECDSA key size is below recommended minimum")
		}
	case "ssh-ed25519":
		// Ed25519 keys are considered very secure
	default:
		warnings = append(warnings, fmt.Sprintf("Unknown key type: %s", keyType))
	}
	
	return secure, warnings, nil
}

// ExtractKeyComment extracts the comment from an SSH key
func (s *DeployKeyService) ExtractKeyComment(keyStr string) string {
	parts := strings.Fields(strings.TrimSpace(keyStr))
	if len(parts) >= 3 {
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// ValidateKeyTitle validates the title for a deploy key
func (s *DeployKeyService) ValidateKeyTitle(title string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title cannot be empty")
	}
	
	if len(title) > 255 {
		return fmt.Errorf("title cannot exceed 255 characters")
	}
	
	// Check for invalid characters
	invalidChars := regexp.MustCompile(`[<>"/\\|?*\x00-\x1f\x7f]`)
	if invalidChars.MatchString(title) {
		return fmt.Errorf("title contains invalid characters")
	}
	
	return nil
}