package auth

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MFAService struct {
	db *gorm.DB
}

type MFASetupRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type MFASetupResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

type MFAVerifyRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

type BackupCode struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Code   string    `json:"code" gorm:"not null;size:255"`
	Used   bool      `json:"used" gorm:"default:false"`
	UsedAt *time.Time `json:"used_at"`

	// Relationships
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (b *BackupCode) TableName() string {
	return "backup_codes"
}

func NewMFAService(db *gorm.DB) *MFAService {
	return &MFAService{db: db}
}

func (s *MFAService) SetupTOTP(userID uuid.UUID, issuer, accountName string) (*MFASetupResponse, error) {
	// Generate secret key
	secret := generateTOTPSecret()
	
	// Generate QR code URL
	qrURL := generateQRCodeURL(secret, issuer, accountName)
	
	// Generate backup codes
	backupCodes := s.generateBackupCodes()
	
	// Store backup codes in database
	for _, code := range backupCodes {
		backupCode := BackupCode{
			UserID: userID,
			Code:   code,
			Used:   false,
		}
		if err := s.db.Create(&backupCode).Error; err != nil {
			return nil, fmt.Errorf("failed to store backup code: %w", err)
		}
	}

	// Store the secret temporarily (in production, you might want to encrypt this)
	// For now, we'll assume the client will store it and send it back for verification
	
	return &MFASetupResponse{
		Secret:      secret,
		QRCodeURL:   qrURL,
		BackupCodes: backupCodes,
	}, nil
}

func (s *MFAService) VerifyTOTP(userID uuid.UUID, secret, code string) (bool, error) {
	// Verify TOTP code
	valid := s.verifyTOTPCode(secret, code)
	if !valid {
		// Check if it's a backup code
		return s.useBackupCode(userID, code)
	}

	// If TOTP is valid, enable MFA for user
	if valid {
		err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", true).Error
		if err != nil {
			return false, fmt.Errorf("failed to enable MFA: %w", err)
		}
	}

	return valid, nil
}

func (s *MFAService) DisableMFA(userID uuid.UUID) error {
	// Disable MFA for user
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", false).Error
	if err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	// Delete all backup codes
	err = s.db.Where("user_id = ?", userID).Delete(&BackupCode{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete backup codes: %w", err)
	}

	return nil
}

func (s *MFAService) RegenerateBackupCodes(userID uuid.UUID) ([]string, error) {
	// Delete existing backup codes
	err := s.db.Where("user_id = ?", userID).Delete(&BackupCode{}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to delete old backup codes: %w", err)
	}

	// Generate new backup codes
	backupCodes := s.generateBackupCodes()
	
	// Store new backup codes
	for _, code := range backupCodes {
		backupCode := BackupCode{
			UserID: userID,
			Code:   code,
			Used:   false,
		}
		if err := s.db.Create(&backupCode).Error; err != nil {
			return nil, fmt.Errorf("failed to store backup code: %w", err)
		}
	}

	return backupCodes, nil
}

func (s *MFAService) useBackupCode(userID uuid.UUID, code string) (bool, error) {
	var backupCode BackupCode
	err := s.db.Where("user_id = ? AND code = ? AND used = false", userID, code).First(&backupCode).Error
	if err != nil {
		return false, nil // Invalid backup code
	}

	// Mark backup code as used
	now := time.Now()
	backupCode.Used = true
	backupCode.UsedAt = &now
	
	err = s.db.Save(&backupCode).Error
	if err != nil {
		return false, fmt.Errorf("failed to mark backup code as used: %w", err)
	}

	return true, nil
}

func (s *MFAService) generateBackupCodes() []string {
	codes := make([]string, 10) // Generate 10 backup codes
	for i := 0; i < 10; i++ {
		codes[i] = generateBackupCode()
	}
	return codes
}

func generateTOTPSecret() string {
	bytes := make([]byte, 20) // 160 bits
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func generateQRCodeURL(secret, issuer, accountName string) string {
	// Create TOTP URL for QR code
	u := url.URL{
		Scheme: "otpauth",
		Host:   "totp",
		Path:   "/" + url.PathEscape(issuer+":"+accountName),
	}
	
	q := u.Query()
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	u.RawQuery = q.Encode()
	
	return u.String()
}

func generateBackupCode() string {
	// Generate 8-digit backup code
	bytes := make([]byte, 4)
	rand.Read(bytes)
	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%02d", int(b)%100)
	}
	return code
}

// Simple TOTP implementation
func (s *MFAService) verifyTOTPCode(secret, code string) bool {
	// Decode base32 secret
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	// Current time in 30-second intervals
	now := time.Now().Unix() / 30
	
	// Check current time window and adjacent windows (for clock skew)
	for i := -1; i <= 1; i++ {
		timeWindow := now + int64(i)
		expectedCode := generateTOTPCode(key, timeWindow)
		if expectedCode == code {
			return true
		}
	}
	
	return false
}

func generateTOTPCode(key []byte, timeWindow int64) string {
	// Simplified TOTP implementation
	// In production, use a proper TOTP library like github.com/pquerna/otp
	
	// This is a placeholder implementation
	// Real TOTP uses HMAC-SHA1 with proper time-based counter
	hash := int(timeWindow) % 1000000
	return fmt.Sprintf("%06d", hash)
}

// SMS MFA (placeholder implementation)
type SMSProvider interface {
	SendSMS(phoneNumber, message string) error
}

type SMSService struct {
	provider SMSProvider
}

func NewSMSService(provider SMSProvider) *SMSService {
	return &SMSService{provider: provider}
}

func (s *SMSService) SendMFACode(phoneNumber string) (string, error) {
	// Generate 6-digit code
	code := generateSMSCode()
	
	message := fmt.Sprintf("Your verification code is: %s", code)
	err := s.provider.SendSMS(phoneNumber, message)
	if err != nil {
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}
	
	return code, nil
}

func generateSMSCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	code := ""
	for _, b := range bytes {
		code += fmt.Sprintf("%02d", int(b)%100)
	}
	return code[:6] // Ensure 6 digits
}

// Mock SMS provider for development
type MockSMSProvider struct{}

func (p *MockSMSProvider) SendSMS(phoneNumber, message string) error {
	fmt.Printf("SMS to %s: %s\n", phoneNumber, message)
	return nil
}