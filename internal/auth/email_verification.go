package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailVerificationToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string     `json:"-" gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	Used      bool       `json:"used" gorm:"default:false"`
	UsedAt    *time.Time `json:"used_at"`

	// Relationships
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

type EmailVerificationService struct {
	db           *gorm.DB
	emailService EmailService
}

func NewEmailVerificationService(db *gorm.DB, emailService EmailService) *EmailVerificationService {
	return &EmailVerificationService{
		db:           db,
		emailService: emailService,
	}
}

func (s *EmailVerificationService) CreateVerificationToken(userID uuid.UUID) (*EmailVerificationToken, error) {
	// Generate secure token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Set expiration time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create token record
	verificationToken := &EmailVerificationToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := s.db.Create(verificationToken).Error; err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	return verificationToken, nil
}

func (s *EmailVerificationService) SendVerificationEmail(userID uuid.UUID) error {
	// Get user
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already verified
	if user.EmailVerified {
		return errors.New("email already verified")
	}

	// Revoke any existing tokens for this user
	s.RevokeUserTokens(userID)

	// Create new verification token
	token, err := s.CreateVerificationToken(userID)
	if err != nil {
		return err
	}

	// Send verification email
	return s.emailService.SendEmailVerification(user.Email, token.Token)
}

func (s *EmailVerificationService) VerifyEmail(token string) error {
	// Find and validate token
	var verificationToken EmailVerificationToken
	err := s.db.Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).First(&verificationToken).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid or expired verification token")
		}
		return fmt.Errorf("database error: %w", err)
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Mark user as verified
		err = tx.Model(&models.User{}).Where("id = ?", verificationToken.UserID).Update("email_verified", true).Error
		if err != nil {
			return fmt.Errorf("failed to verify user: %w", err)
		}

		// Mark token as used
		now := time.Now()
		verificationToken.Used = true
		verificationToken.UsedAt = &now
		err = tx.Save(&verificationToken).Error
		if err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}

		return nil
	})
}

func (s *EmailVerificationService) IsEmailVerified(userID uuid.UUID) (bool, error) {
	var user models.User
	err := s.db.Select("email_verified").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return false, err
	}
	return user.EmailVerified, nil
}

func (s *EmailVerificationService) RevokeUserTokens(userID uuid.UUID) error {
	// Mark all active tokens for user as used
	now := time.Now()
	err := s.db.Model(&EmailVerificationToken{}).
		Where("user_id = ? AND used = false AND expires_at > ?", userID, time.Now()).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": &now,
		}).Error

	return err
}

func (s *EmailVerificationService) CleanupExpiredTokens() error {
	// Delete expired tokens older than 7 days
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	result := s.db.Where("expires_at < ?", cutoff).Delete(&EmailVerificationToken{})
	return result.Error
}

func (s *EmailVerificationService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Helper method to get verification status and token info
func (s *EmailVerificationService) GetVerificationStatus(userID uuid.UUID) (bool, *EmailVerificationToken, error) {
	var user models.User
	err := s.db.Select("email_verified").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return false, nil, err
	}

	if user.EmailVerified {
		return true, nil, nil
	}

	// Get active token if any
	var token EmailVerificationToken
	err = s.db.Where("user_id = ? AND used = false AND expires_at > ?", userID, time.Now()).
		Order("created_at desc").
		First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil, nil // Not verified, no active token
	}

	if err != nil {
		return false, nil, err
	}

	return false, &token, nil
}
