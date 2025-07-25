package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SecretService handles secret management operations
type SecretService struct {
	db         *gorm.DB
	logger     *logrus.Logger
	encryptKey []byte // 32-byte key for AES-256
}

// NewSecretService creates a new secret service
func NewSecretService(db *gorm.DB, logger *logrus.Logger, encryptKey string) *SecretService {
	// Ensure we have a 32-byte key for AES-256
	key := make([]byte, 32)
	copy(key, []byte(encryptKey))
	
	return &SecretService{
		db:         db,
		logger:     logger,
		encryptKey: key,
	}
}

// encrypt encrypts a value using AES-256-GCM
func (s *SecretService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a value using AES-256-GCM
func (s *SecretService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext2 := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext2, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ListSecrets lists secrets for a repository or organization
func (s *SecretService) ListSecrets(ctx context.Context, repositoryID *uuid.UUID, organizationID *uuid.UUID) ([]models.Secret, error) {
	query := s.db.WithContext(ctx).Model(&models.Secret{})
	
	if repositoryID != nil {
		query = query.Where("repository_id = ?", *repositoryID)
	} else if organizationID != nil {
		query = query.Where("organization_id = ?", *organizationID)
	} else {
		return nil, fmt.Errorf("either repository_id or organization_id must be provided")
	}

	var secrets []models.Secret
	err := query.Order("created_at DESC").Find(&secrets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	// Note: We don't return the encrypted values in the list
	return secrets, nil
}

// CreateSecret creates a new secret
func (s *SecretService) CreateSecret(ctx context.Context, repositoryID *uuid.UUID, organizationID *uuid.UUID, name, value string, environment *string) (*models.Secret, error) {
	// Encrypt the value
	encryptedValue, err := s.encrypt(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
	}

	secret := &models.Secret{
		Name:           name,
		EncryptedValue: encryptedValue,
		RepositoryID:   repositoryID,
		OrganizationID: organizationID,
		Environment:    environment,
	}

	if err := s.db.WithContext(ctx).Create(secret).Error; err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"secret_id":       secret.ID,
		"repository_id":   repositoryID,
		"organization_id": organizationID,
		"name":            name,
		"environment":     environment,
	}).Info("Created new secret")

	return secret, nil
}

// GetSecret retrieves a secret by ID
func (s *SecretService) GetSecret(ctx context.Context, id uuid.UUID) (*models.Secret, error) {
	var secret models.Secret
	err := s.db.WithContext(ctx).First(&secret, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return &secret, nil
}

// GetSecretValue retrieves and decrypts a secret value
func (s *SecretService) GetSecretValue(ctx context.Context, id uuid.UUID) (string, error) {
	secret, err := s.GetSecret(ctx, id)
	if err != nil {
		return "", err
	}

	value, err := s.decrypt(secret.EncryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret value: %w", err)
	}

	return value, nil
}

// UpdateSecret updates a secret's value
func (s *SecretService) UpdateSecret(ctx context.Context, id uuid.UUID, value string, environment *string) (*models.Secret, error) {
	secret, err := s.GetSecret(ctx, id)
	if err != nil {
		return nil, err
	}

	// Encrypt the new value
	encryptedValue, err := s.encrypt(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret value: %w", err)
	}

	updates := map[string]interface{}{
		"encrypted_value": encryptedValue,
	}

	if environment != nil {
		updates["environment"] = *environment
	}

	if err := s.db.WithContext(ctx).Model(secret).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update secret: %w", err)
	}

	// Reload secret with updates
	return s.GetSecret(ctx, id)
}

// DeleteSecret deletes a secret
func (s *SecretService) DeleteSecret(ctx context.Context, id uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.Secret{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete secret: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("secret not found")
	}

	s.logger.WithField("secret_id", id).Info("Deleted secret")
	return nil
}

// GetSecretsForJob retrieves all secrets available for a job execution
func (s *SecretService) GetSecretsForJob(ctx context.Context, repositoryID uuid.UUID, organizationID *uuid.UUID, environment *string) (map[string]string, error) {
	query := s.db.WithContext(ctx).Model(&models.Secret{})
	
	// Build query to get repository and organization secrets
	var conditions []string
	var args []interface{}
	
	// Repository secrets
	conditions = append(conditions, "repository_id = ?")
	args = append(args, repositoryID)
	
	// Organization secrets (if available)
	if organizationID != nil {
		conditions = append(conditions, "organization_id = ?")
		args = append(args, *organizationID)
	}
	
	// Environment filtering
	if environment != nil {
		query = query.Where("environment IS NULL OR environment = ?", *environment)
	} else {
		query = query.Where("environment IS NULL")
	}
	
	// Combine all conditions with OR
	whereClause := "(" + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " OR " + conditions[i]
	}
	whereClause += ")"
	
	var secrets []models.Secret
	err := query.Where(whereClause, args...).Find(&secrets).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets for job: %w", err)
	}

	// Decrypt all secrets and return as map
	secretMap := make(map[string]string)
	for _, secret := range secrets {
		value, err := s.decrypt(secret.EncryptedValue)
		if err != nil {
			s.logger.WithError(err).WithField("secret_id", secret.ID).Error("Failed to decrypt secret for job")
			continue // Skip this secret but continue with others
		}
		secretMap[secret.Name] = value
	}

	return secretMap, nil
}