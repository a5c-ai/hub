package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrAccountLocked      = errors.New("account is locked")
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	MFACode  string `json:"mfa_code,omitempty"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	FullName string `json:"full_name" binding:"required,min=1,max=255"`
}

type AuthResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=12"`
}

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Register(ctx context.Context, req RegisterRequest) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
	VerifyToken(ctx context.Context, token string) (*models.User, error)
	RequestPasswordReset(ctx context.Context, req PasswordResetRequest) error
	ResetPassword(ctx context.Context, req PasswordResetConfirmRequest) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error
	// Legacy methods for backward compatibility
	GetUserByID(userID uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	UpdateUser(user *models.User) error
	ValidateToken(tokenString string) (*models.User, error)
}

type authService struct {
	db             *gorm.DB
	jwtManager     *JWTManager
	config         *config.Config
	sessionService *SessionService
	blacklistService *TokenBlacklistService
}

func NewAuthService(db *gorm.DB, jwtManager *JWTManager, cfg *config.Config) AuthService {
	sessionService := NewSessionService(db)
	blacklistService := NewTokenBlacklistService(db)
	
	return &authService{
		db:               db,
		jwtManager:       jwtManager,
		config:           cfg,
		sessionService:   sessionService,
		blacklistService: blacklistService,
	}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user models.User
	
	// Support login with either email or username
	var err error
	if req.Email != "" {
		err = s.db.Where("email = ?", req.Email).First(&user).Error
	} else {
		err = s.db.Where("username = ? OR email = ?", req.Email, req.Email).First(&user).Error
	}
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if account is active
	if !user.IsActive {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check MFA if enabled
	if user.TwoFactorEnabled {
		if req.MFACode == "" {
			return nil, errors.New("MFA code required")
		}
		
		// Verify MFA code using MFA service
		mfaService := NewMFAService(s.db)
		valid, err := mfaService.VerifyMFACode(user.ID, req.MFACode)
		if err != nil {
			return nil, fmt.Errorf("MFA verification failed: %w", err)
		}
		if !valid {
			return nil, errors.New("invalid MFA code")
		}
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Create a proper session with refresh token
	session, err := s.sessionService.CreateSession(user.ID, "", "", false)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// Remove sensitive information before returning
	user.PasswordHash = ""

	return &AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    int64(time.Duration(s.config.JWT.ExpirationHour) * time.Hour / time.Second),
	}, nil
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	err := s.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error
	if err == nil {
		return nil, ErrUserExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	user := models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		IsActive:     true,
		IsAdmin:      false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email
	emailService := NewSMTPEmailService(s.config)
	verificationService := NewEmailVerificationService(s.db, emailService)
	if err := verificationService.SendVerificationEmail(user.ID); err != nil {
		// Log the error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	// Remove sensitive information before returning
	user.PasswordHash = ""

	return &user, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate and refresh the session
	session, err := s.sessionService.RefreshSession(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	// Get user information
	var user models.User
	err = s.db.First(&user, session.UserID).Error
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		// Revoke session for inactive user
		s.sessionService.RevokeSession(refreshToken)
		return nil, errors.New("account is disabled")
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Remove sensitive information
	user.PasswordHash = ""

	return &AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    int64(time.Duration(s.config.JWT.ExpirationHour) * time.Hour / time.Second),
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID) error {
	// Revoke all user sessions
	if err := s.sessionService.RevokeUserSessions(userID); err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	// Add current tokens to blacklist (if any active tokens exist)
	if err := s.blacklistService.BlacklistUserTokens(userID); err != nil {
		return fmt.Errorf("failed to blacklist user tokens: %w", err)
	}

	return nil
}

// LogoutFromDevice revokes a specific session/device
func (s *authService) LogoutFromDevice(ctx context.Context, refreshToken string) error {
	// Revoke specific session
	if err := s.sessionService.RevokeSession(refreshToken); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Blacklist the refresh token
	if err := s.blacklistService.BlacklistToken(refreshToken, time.Now().Add(30*24*time.Hour)); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

func (s *authService) VerifyToken(ctx context.Context, token string) (*models.User, error) {
	// Check if token is blacklisted
	if blacklisted, err := s.blacklistService.IsTokenBlacklisted(token); err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	} else if blacklisted {
		return nil, errors.New("token has been revoked")
	}

	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.First(&user, claims.UserID).Error
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("account is disabled")
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, req PasswordResetRequest) error {
	var user models.User
	err := s.db.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Initialize password reset service
	passwordResetService := NewPasswordResetService(s.db)
	emailService := NewSMTPEmailService(s.config)

	// Generate reset token
	resetToken, err := passwordResetService.CreateResetToken(user.ID)
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}
	
	// Send reset email
	if err := emailService.SendPasswordResetEmail(user.Email, resetToken.Token); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

func (s *authService) ResetPassword(ctx context.Context, req PasswordResetConfirmRequest) error {
	// Initialize password reset service
	passwordResetService := NewPasswordResetService(s.db)
	
	// Use the reset token to change password
	err := passwordResetService.UseResetToken(req.Token, req.Password)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) error {
	emailService := NewSMTPEmailService(s.config)
	verificationService := NewEmailVerificationService(s.db, emailService)
	return verificationService.VerifyEmail(token)
}

func (s *authService) ResendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	emailService := NewSMTPEmailService(s.config)
	verificationService := NewEmailVerificationService(s.db, emailService)
	return verificationService.SendVerificationEmail(userID)
}

// Legacy methods for backward compatibility
func (s *authService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *authService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *authService) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("username = ? AND is_active = ?", username, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *authService) UpdateUser(user *models.User) error {
	// Don't allow updating sensitive fields through this method
	updates := map[string]interface{}{
		"full_name":  user.FullName,
		"bio":        user.Bio,
		"company":    user.Company,
		"location":   user.Location,
		"website":    user.Website,
		"avatar_url": user.AvatarURL,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(user).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *authService) ValidateToken(tokenString string) (*models.User, error) {
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	return s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error
}

func (s *authService) InitiatePasswordReset(email string) error {
	user, err := s.GetUserByEmail(email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// TODO: Generate password reset token and send email
	// For now, just log that password reset was requested
	_ = user
	return nil
}

// GetUserSessions returns active sessions for a user
func (s *authService) GetUserSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	return s.sessionService.GetUserSessions(userID)
}

// RevokeUserSession revokes a specific user session
func (s *authService) RevokeUserSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	// Get session to verify it belongs to the user
	sessions, err := s.sessionService.GetUserSessions(userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.ID == sessionID {
			return s.sessionService.RevokeSession(session.RefreshToken)
		}
	}

	return errors.New("session not found")
}

func (s *authService) sendVerificationEmail(user *models.User) error {
	// TODO: Implement email sending
	fmt.Printf("Sending verification email to %s\n", user.Email)
	return nil
}

func (s *authService) sendPasswordResetEmail(user *models.User, token string) error {
	// TODO: Implement email sending
	fmt.Printf("Sending password reset email to %s with token %s\n", user.Email, token)
	return nil
}

func generateAuthSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}