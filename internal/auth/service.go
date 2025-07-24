package auth

import (
	"errors"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
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

type AuthService struct {
	db         *db.Database
	jwtManager *JWTManager
	config     *config.Config
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=12"`
	FullName string `json:"full_name" binding:"required,min=1,max=255"`
}

type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

type PasswordResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type PasswordResetConfirmRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=12"`
}

func NewAuthService(database *db.Database, cfg *config.Config) *AuthService {
	jwtManager := NewJWTManager(cfg.JWT)
	return &AuthService{
		db:         database,
		jwtManager: jwtManager,
		config:     cfg,
	}
}

func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	var user models.User
	if err := s.db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if account is active
	if !user.IsActive {
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.DB.Save(&user)

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	// Remove sensitive information before returning
	user.PasswordHash = ""

	return &AuthResponse{
		User:  &user,
		Token: token,
	}, nil
}

func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
		return nil, ErrUserExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
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

	if err := s.db.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	// Remove sensitive information before returning
	user.PasswordHash = ""

	return &AuthResponse{
		User:  &user,
		Token: token,
	}, nil
}

func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *AuthService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.DB.Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove sensitive information
	user.PasswordHash = ""
	return &user, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*models.User, error) {
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

func (s *AuthService) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	var user models.User
	if err := s.db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
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
	return s.db.DB.Model(&user).Update("password_hash", string(hashedPassword)).Error
}

func (s *AuthService) InitiatePasswordReset(email string) error {
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

func (s *AuthService) ResetPassword(token, newPassword string) error {
	// TODO: Implement password reset token validation and password update
	return errors.New("password reset not implemented yet")
}

func (s *AuthService) VerifyEmail(token string) error {
	// TODO: Implement email verification
	return errors.New("email verification not implemented yet")
}