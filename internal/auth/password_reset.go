package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
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

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

type PasswordResetService struct {
	db *gorm.DB
}

func NewPasswordResetService(db *gorm.DB) *PasswordResetService {
	return &PasswordResetService{db: db}
}

func (s *PasswordResetService) CreateResetToken(userID uuid.UUID) (*PasswordResetToken, error) {
	// Generate secure token
	token, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Set expiration time (1 hour from now)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create token record
	resetToken := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		return nil, fmt.Errorf("failed to create reset token: %w", err)
	}

	return resetToken, nil
}

func (s *PasswordResetService) ValidateResetToken(token string) (*PasswordResetToken, error) {
	var resetToken PasswordResetToken
	err := s.db.Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).First(&resetToken).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired reset token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &resetToken, nil
}

func (s *PasswordResetService) UseResetToken(token string, newPassword string) error {
	// Validate token
	resetToken, err := s.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// Start transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Hash new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Update user password
		err = tx.Model(&models.User{}).Where("id = ?", resetToken.UserID).Update("password_hash", string(hashedPassword)).Error
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		// Mark token as used
		now := time.Now()
		resetToken.Used = true
		resetToken.UsedAt = &now
		err = tx.Save(resetToken).Error
		if err != nil {
			return fmt.Errorf("failed to mark token as used: %w", err)
		}

		return nil
	})
}

func (s *PasswordResetService) CleanupExpiredTokens() error {
	// Delete expired tokens older than 24 hours
	result := s.db.Where("expires_at < ?", time.Now().Add(-24*time.Hour)).Delete(&PasswordResetToken{})
	return result.Error
}

func (s *PasswordResetService) RevokeUserTokens(userID uuid.UUID) error {
	// Mark all active tokens for user as used
	now := time.Now()
	err := s.db.Model(&PasswordResetToken{}).
		Where("user_id = ? AND used = false AND expires_at > ?", userID, time.Now()).
		Updates(map[string]interface{}{
			"used":    true,
			"used_at": &now,
		}).Error
	
	return err
}

func (s *PasswordResetService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Email service interface for sending password reset emails
type EmailService interface {
	SendPasswordResetEmail(to, token string) error
	SendEmailVerification(to, token string) error
	SendMFASetupEmail(to string, backupCodes []string) error
}

// Mock email service for development
type MockEmailService struct{}

func (s *MockEmailService) SendPasswordResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	fmt.Printf("Password Reset Email to %s:\nReset your password: %s\n", to, resetURL)
	return nil
}

func (s *MockEmailService) SendEmailVerification(to, token string) error {
	verifyURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
	fmt.Printf("Email Verification to %s:\nVerify your email: %s\n", to, verifyURL)
	return nil
}

func (s *MockEmailService) SendMFASetupEmail(to string, backupCodes []string) error {
	fmt.Printf("MFA Setup Email to %s:\nBackup codes: %v\n", to, backupCodes)
	return nil
}