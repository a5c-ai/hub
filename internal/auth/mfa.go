package auth

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"gorm.io/gorm"
)

type MFAService struct {
	db           *gorm.DB
	emailService EmailService
}

type MFASetupRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type MFASetupResponse struct {
	Secret      string   `json:"secret"`
	QRCodeURL   string   `json:"qr_code_url"`
	BackupCodes []string `json:"backup_codes"`
}

type MFAVerifyRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Code   string    `json:"code"`
}

type BackupCode struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Code   string     `json:"code" gorm:"not null;size:255"`
	Used   bool       `json:"used" gorm:"default:false"`
	UsedAt *time.Time `json:"used_at"`

	// Relationships
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type WebAuthnCredential struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID       uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	CredentialID string     `json:"credential_id" gorm:"not null;uniqueIndex;size:255"`
	PublicKey    []byte     `json:"public_key" gorm:"not null"`
	Name         string     `json:"name" gorm:"not null;size:255"`
	SignCount    uint32     `json:"sign_count" gorm:"default:0"`
	LastUsedAt   *time.Time `json:"last_used_at"`

	// Relationships
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type SMSVerificationCode struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Code      string     `json:"code" gorm:"not null;size:10"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	Used      bool       `json:"used" gorm:"default:false"`
	UsedAt    *time.Time `json:"used_at"`

	// Relationships
	User models.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (b *BackupCode) TableName() string {
	return "backup_codes"
}

func (w *WebAuthnCredential) TableName() string {
	return "webauthn_credentials"
}

func (s *SMSVerificationCode) TableName() string {
	return "sms_verification_codes"
}

func NewMFAService(db *gorm.DB) *MFAService {
	return &MFAService{
		db:           db,
		emailService: nil, // Will be set separately when needed
	}
}

func NewMFAServiceWithEmail(db *gorm.DB, emailService EmailService) *MFAService {
	return &MFAService{
		db:           db,
		emailService: emailService,
	}
}

func (s *MFAService) SetupTOTP(userID uuid.UUID, issuer, accountName string) (*MFASetupResponse, error) {
	// Generate TOTP key using proper library
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: accountName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate QR code URL from the key
	qrURL := key.URL()

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
		Secret:      key.Secret(),
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

	// If TOTP is valid, enable MFA for user and store the secret
	if valid {
		err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
			"two_factor_enabled": true,
			"two_factor_secret":  secret,
		}).Error
		if err != nil {
			return false, fmt.Errorf("failed to enable MFA: %w", err)
		}

		// Send MFA setup notification email
		if s.emailService != nil {
			var user models.User
			if err := s.db.Where("id = ?", userID).First(&user).Error; err == nil {
				// Get backup codes for this user
				var backupCodes []BackupCode
				s.db.Where("user_id = ? AND used = false", userID).Find(&backupCodes)

				codes := make([]string, len(backupCodes))
				for i, bc := range backupCodes {
					codes[i] = bc.Code
				}

				// Send email notification (don't fail if email fails)
				if err := s.emailService.SendMFASetupEmail(user.Email, codes); err != nil {
					fmt.Printf("Failed to send MFA setup email: %v\n", err)
				}
			}
		}
	}

	return valid, nil
}

// VerifyMFACode verifies any type of MFA code for login
func (s *MFAService) VerifyMFACode(userID uuid.UUID, code string) (bool, error) {
	// Get user with MFA settings
	var user models.User
	err := s.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	if !user.TwoFactorEnabled {
		return false, errors.New("MFA not enabled for user")
	}

	// Try TOTP first if secret is available
	if user.TwoFactorSecret != "" {
		if s.verifyTOTPCode(user.TwoFactorSecret, code) {
			return true, nil
		}
	}

	// Try SMS verification code
	if s.verifySMSCode(userID, code) {
		return true, nil
	}

	// Try backup code as last resort
	return s.useBackupCode(userID, code)
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

	// Send email notification about new backup codes
	if s.emailService != nil {
		var user models.User
		if err := s.db.Where("id = ?", userID).First(&user).Error; err == nil {
			// Send email notification (don't fail if email fails)
			if err := s.emailService.SendMFASetupEmail(user.Email, backupCodes); err != nil {
				fmt.Printf("Failed to send backup codes regeneration email: %v\n", err)
			}
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

// TOTP implementation using proper library
func (s *MFAService) verifyTOTPCode(secret, code string) bool {
	// Use the proper TOTP library for validation
	valid := totp.Validate(code, secret)
	return valid
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
	if len(code) > 6 {
		return code[:6] // Ensure 6 digits
	}
	// Pad with zeros if needed
	for len(code) < 6 {
		code = "0" + code
	}
	return code
}

// LoggingSMSProvider logs SMS messages when no real provider is configured
type LoggingSMSProvider struct{}

func (p *LoggingSMSProvider) SendSMS(phoneNumber, message string) error {
	fmt.Printf("=== SMS LOG (No SMS provider configured) ===\n")
	fmt.Printf("To: %s\n", phoneNumber)
	fmt.Printf("Message: %s\n", message)
	fmt.Printf("==========================================\n")

	// In production, you might want to:
	// 1. Use a real SMS provider (Twilio, AWS SNS, etc.)
	// 2. Store in database for audit trail
	// 3. Send to a message queue for later processing
	// 4. Use alternative notification methods

	return nil
}

// SMS MFA methods
func (s *MFAService) SendSMSCode(userID uuid.UUID, phoneNumber string) error {
	// Generate 6-digit code
	code := generateSMSCode()

	// Store code in database with expiration
	expiresAt := time.Now().Add(5 * time.Minute)
	smsCode := SMSVerificationCode{
		UserID:    userID,
		Code:      code,
		ExpiresAt: expiresAt,
		Used:      false,
	}

	if err := s.db.Create(&smsCode).Error; err != nil {
		return fmt.Errorf("failed to store SMS code: %w", err)
	}

	// Send SMS (using logging provider when no real provider configured)
	provider := &LoggingSMSProvider{}
	message := fmt.Sprintf("Your verification code is: %s. Valid for 5 minutes.", code)
	return provider.SendSMS(phoneNumber, message)
}

func (s *MFAService) verifySMSCode(userID uuid.UUID, code string) bool {
	var smsCode SMSVerificationCode
	err := s.db.Where("user_id = ? AND code = ? AND used = false AND expires_at > ?",
		userID, code, time.Now()).First(&smsCode).Error

	if err != nil {
		return false
	}

	// Mark code as used
	now := time.Now()
	smsCode.Used = true
	smsCode.UsedAt = &now
	s.db.Save(&smsCode)

	return true
}

// WebAuthn methods (basic implementation)
type WebAuthnRegistrationRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

type WebAuthnRegistrationResponse struct {
	Options string `json:"options"` // JSON string of WebAuthn creation options
}

type WebAuthnLoginRequest struct {
	UserID uuid.UUID `json:"user_id"`
}

type WebAuthnLoginResponse struct {
	Options string `json:"options"` // JSON string of WebAuthn assertion options
}

func (s *MFAService) InitiateWebAuthnRegistration(userID uuid.UUID, credentialName string) (*WebAuthnRegistrationResponse, error) {
	// This is a simplified implementation
	// In production, you would use a proper WebAuthn library like github.com/go-webauthn/webauthn

	// For now, return mock options
	options := map[string]interface{}{
		"challenge": generateMFASecureToken(),
		"rp": map[string]interface{}{
			"name": "A5C Hub",
			"id":   "localhost",
		},
		"user": map[string]interface{}{
			"id":          userID.String(),
			"name":        credentialName,
			"displayName": credentialName,
		},
		"pubKeyCredParams": []map[string]interface{}{
			{"alg": -7, "type": "public-key"},
			{"alg": -257, "type": "public-key"},
		},
		"authenticatorSelection": map[string]interface{}{
			"authenticatorAttachment": "platform",
			"userVerification":        "required",
		},
		"timeout": 60000,
	}

	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	return &WebAuthnRegistrationResponse{
		Options: string(optionsJSON),
	}, nil
}

func (s *MFAService) CompleteWebAuthnRegistration(userID uuid.UUID, credentialName string, publicKeyBytes []byte, credentialID string) error {
	// Store the WebAuthn credential
	credential := WebAuthnCredential{
		UserID:       userID,
		CredentialID: credentialID,
		PublicKey:    publicKeyBytes,
		Name:         credentialName,
		SignCount:    0,
	}

	if err := s.db.Create(&credential).Error; err != nil {
		return fmt.Errorf("failed to store WebAuthn credential: %w", err)
	}

	// Enable MFA for user if not already enabled
	err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("two_factor_enabled", true).Error
	if err != nil {
		return fmt.Errorf("failed to enable MFA: %w", err)
	}

	return nil
}

func (s *MFAService) GetWebAuthnCredentials(userID uuid.UUID) ([]WebAuthnCredential, error) {
	var credentials []WebAuthnCredential
	err := s.db.Where("user_id = ?", userID).Find(&credentials).Error
	return credentials, err
}

func (s *MFAService) DeleteWebAuthnCredential(userID uuid.UUID, credentialID string) error {
	return s.db.Where("user_id = ? AND credential_id = ?", userID, credentialID).Delete(&WebAuthnCredential{}).Error
}

func generateMFASecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
