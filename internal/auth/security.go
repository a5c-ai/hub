package auth

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Security event types
const (
	EventLogin          = "login"
	EventLoginFailed    = "login_failed"
	EventPasswordReset  = "password_reset"
	EventMFAEnabled     = "mfa_enabled"
	EventMFADisabled    = "mfa_disabled"
	EventAccountLocked  = "account_locked"
	EventOAuthLogin     = "oauth_login"
)

type SecurityEvent struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	EventType   string     `json:"event_type" gorm:"not null;size:50;index"`
	IPAddress   string     `json:"ip_address" gorm:"size:45;index"`
	UserAgent   string     `json:"user_agent" gorm:"size:255"`
	Details     string     `json:"details" gorm:"type:text"`
	Severity    string     `json:"severity" gorm:"size:20;default:'info'"` // info, warning, critical
}

func (SecurityEvent) TableName() string {
	return "security_events"
}

type LoginAttempt struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID      *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	Email       string     `json:"email" gorm:"not null;size:255;index"`
	IPAddress   string     `json:"ip_address" gorm:"not null;size:45;index"`
	Success     bool       `json:"success" gorm:"not null;index"`
	UserAgent   string     `json:"user_agent" gorm:"size:255"`
	FailReason  string     `json:"fail_reason" gorm:"size:255"`
}

func (LoginAttempt) TableName() string {
	return "login_attempts"
}

type SecurityService struct {
	db *gorm.DB
}

func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{db: db}
}

// Rate limiting configuration
type RateLimitConfig struct {
	MaxAttempts     int           // Maximum failed attempts
	WindowDuration  time.Duration // Time window for rate limiting
	LockoutDuration time.Duration // How long to lock account
}

var DefaultRateLimitConfig = RateLimitConfig{
	MaxAttempts:     5,
	WindowDuration:  15 * time.Minute,
	LockoutDuration: 30 * time.Minute,
}

func (s *SecurityService) RecordLoginAttempt(userID *uuid.UUID, email, ipAddress, userAgent string, success bool, failReason string) error {
	attempt := &LoginAttempt{
		UserID:     userID,
		Email:      email,
		IPAddress:  ipAddress,
		Success:    success,
		UserAgent:  userAgent,
		FailReason: failReason,
	}

	if err := s.db.Create(attempt).Error; err != nil {
		return fmt.Errorf("failed to record login attempt: %w", err)
	}

	// Record security event
	eventType := EventLogin
	severity := "info"
	if !success {
		eventType = EventLoginFailed
		severity = "warning"
	}

	return s.RecordSecurityEvent(userID, eventType, ipAddress, userAgent, failReason, severity)
}

func (s *SecurityService) CheckRateLimit(email, ipAddress string, config RateLimitConfig) error {
	now := time.Now()
	windowStart := now.Add(-config.WindowDuration)

	// Check failed attempts by email
	var emailFailures int64
	s.db.Model(&LoginAttempt{}).
		Where("email = ? AND success = false AND created_at > ?", email, windowStart).
		Count(&emailFailures)

	// Check failed attempts by IP address
	var ipFailures int64
	s.db.Model(&LoginAttempt{}).
		Where("ip_address = ? AND success = false AND created_at > ?", ipAddress, windowStart).
		Count(&ipFailures)

	// If either email or IP has too many failures, apply rate limit
	if emailFailures >= int64(config.MaxAttempts) {
		return fmt.Errorf("too many failed login attempts for this email. Try again in %v", config.LockoutDuration)
	}

	if ipFailures >= int64(config.MaxAttempts) {
		return fmt.Errorf("too many failed login attempts from this IP address. Try again in %v", config.LockoutDuration)
	}

	return nil
}

func (s *SecurityService) IsAccountLocked(email string, config RateLimitConfig) (bool, time.Time) {
	now := time.Now()
	windowStart := now.Add(-config.LockoutDuration)

	var lastFailure LoginAttempt
	err := s.db.Where("email = ? AND success = false AND created_at > ?", email, windowStart).
		Order("created_at desc").
		First(&lastFailure).Error

	if err != nil {
		return false, time.Time{} // No recent failures
	}

	// Check if we have enough failures to trigger lockout
	var failures int64
	s.db.Model(&LoginAttempt{}).
		Where("email = ? AND success = false AND created_at > ?", email, now.Add(-config.WindowDuration)).
		Count(&failures)

	if failures >= int64(config.MaxAttempts) {
		unlockTime := lastFailure.CreatedAt.Add(config.LockoutDuration)
		return time.Now().Before(unlockTime), unlockTime
	}

	return false, time.Time{}
}

func (s *SecurityService) RecordSecurityEvent(userID *uuid.UUID, eventType, ipAddress, userAgent, details, severity string) error {
	event := &SecurityEvent{
		UserID:    userID,
		EventType: eventType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Severity:  severity,
	}

	return s.db.Create(event).Error
}

func (s *SecurityService) GetSecurityEvents(userID uuid.UUID, limit int) ([]SecurityEvent, error) {
	var events []SecurityEvent
	query := s.db.Where("user_id = ?", userID).Order("created_at desc")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&events).Error
	return events, err
}

func (s *SecurityService) GetSuspiciousActivity(hoursBack int) ([]SecurityEvent, error) {
	var events []SecurityEvent
	since := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	err := s.db.Where("severity IN (?, ?) AND created_at > ?", "warning", "critical", since).
		Order("created_at desc").
		Find(&events).Error
	
	return events, err
}

func (s *SecurityService) ValidateIPAddress(ipStr string) error {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return errors.New("invalid IP address")
	}
	
	// Check for private/local IPs in production
	if s.isPrivateIP(ip) {
		// Log but don't block private IPs for development
		return nil
	}
	
	return nil
}

func (s *SecurityService) isPrivateIP(ip net.IP) bool {
	// Check for private IP ranges
	private := false
	_, cidr1, _ := net.ParseCIDR("10.0.0.0/8")
	_, cidr2, _ := net.ParseCIDR("172.16.0.0/12")
	_, cidr3, _ := net.ParseCIDR("192.168.0.0/16")
	_, cidr4, _ := net.ParseCIDR("127.0.0.0/8")
	
	if cidr1.Contains(ip) || cidr2.Contains(ip) || cidr3.Contains(ip) || cidr4.Contains(ip) {
		private = true
	}
	
	return private
}

func (s *SecurityService) CleanupOldEvents(daysToKeep int) error {
	cutoff := time.Now().Add(-time.Duration(daysToKeep) * 24 * time.Hour)
	
	// Delete old security events
	err := s.db.Where("created_at < ?", cutoff).Delete(&SecurityEvent{}).Error
	if err != nil {
		return fmt.Errorf("failed to cleanup security events: %w", err)
	}
	
	// Delete old login attempts
	err = s.db.Where("created_at < ?", cutoff).Delete(&LoginAttempt{}).Error
	if err != nil {
		return fmt.Errorf("failed to cleanup login attempts: %w", err)
	}
	
	return nil
}

// Password strength validation
func (s *SecurityService) ValidatePasswordStrength(password string) []string {
	var issues []string
	
	if len(password) < 12 {
		issues = append(issues, "Password must be at least 12 characters long")
	}
	
	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
	
	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 126 && !(char >= 'A' && char <= 'Z') && !(char >= 'a' && char <= 'z') && !(char >= '0' && char <= '9'):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		issues = append(issues, "Password must contain at least one uppercase letter")
	}
	if !hasLower {
		issues = append(issues, "Password must contain at least one lowercase letter")
	}
	if !hasDigit {
		issues = append(issues, "Password must contain at least one digit")
	}
	if !hasSpecial {
		issues = append(issues, "Password must contain at least one special character")
	}
	
	return issues
}