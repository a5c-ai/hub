package auth

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
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

// Rate Limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	
	// Get existing requests for this key
	requests := rl.requests[key]
	
	// Remove old requests outside the time window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if now.Sub(reqTime) < rl.window {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	// Check if limit exceeded
	if len(validRequests) >= rl.limit {
		return false
	}
	
	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	
	return true
}

func (rl *RateLimiter) Cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	for key, requests := range rl.requests {
		validRequests := make([]time.Time, 0)
		for _, reqTime := range requests {
			if now.Sub(reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}
		
		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}

// Account lockout functionality
type AccountLockout struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	IPAddress    string    `json:"ip_address" gorm:"size:45;index"`
	Reason       string    `json:"reason" gorm:"size:255"`
	FailedAttempts int     `json:"failed_attempts" gorm:"default:0"`
	LockedUntil  *time.Time `json:"locked_until"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
}

func (AccountLockout) TableName() string {
	return "account_lockouts"
}

type SecurityService struct {
	db                   *gorm.DB
	loginLimiter        *RateLimiter
	registrationLimiter *RateLimiter
	passwordResetLimiter *RateLimiter
	mfaLimiter          *RateLimiter
	generalLimiter      *RateLimiter
}

func NewSecurityService(db *gorm.DB) *SecurityService {
	return &SecurityService{
		db:                   db,
		loginLimiter:        NewRateLimiter(5, 15*time.Minute),   // 5 login attempts per 15 minutes
		registrationLimiter: NewRateLimiter(3, time.Hour),       // 3 registrations per hour
		passwordResetLimiter: NewRateLimiter(3, time.Hour),      // 3 password resets per hour
		mfaLimiter:          NewRateLimiter(10, 5*time.Minute),  // 10 MFA attempts per 5 minutes
		generalLimiter:      NewRateLimiter(100, time.Minute),   // 100 requests per minute
	}
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

// Enhanced rate limiting methods
func (s *SecurityService) CheckLoginRateLimit(ipAddress string) bool {
	return s.loginLimiter.IsAllowed(ipAddress)
}

func (s *SecurityService) CheckRegistrationRateLimit(ipAddress string) bool {
	return s.registrationLimiter.IsAllowed(ipAddress)
}

func (s *SecurityService) CheckPasswordResetRateLimit(ipAddress string) bool {
	return s.passwordResetLimiter.IsAllowed(ipAddress)
}

func (s *SecurityService) CheckMFARateLimit(ipAddress string) bool {
	return s.mfaLimiter.IsAllowed(ipAddress)
}

func (s *SecurityService) CheckGeneralRateLimit(ipAddress string) bool {
	return s.generalLimiter.IsAllowed(ipAddress)
}

// Enhanced audit logging
type AuditEvent string

const (
	AuditEventLogin              AuditEvent = "login"
	AuditEventLoginFailed        AuditEvent = "login_failed"
	AuditEventLogout             AuditEvent = "logout"
	AuditEventRegister           AuditEvent = "register"
	AuditEventPasswordChange     AuditEvent = "password_change"
	AuditEventPasswordReset      AuditEvent = "password_reset"
	AuditEventMFASetup           AuditEvent = "mfa_setup"
	AuditEventMFADisable         AuditEvent = "mfa_disable"
	AuditEventMFAFailed          AuditEvent = "mfa_failed"
	AuditEventOAuthLink          AuditEvent = "oauth_link"
	AuditEventOAuthUnlink        AuditEvent = "oauth_unlink"
	AuditEventSessionRevoked     AuditEvent = "session_revoked"
	AuditEventSuspiciousActivity AuditEvent = "suspicious_activity"
	AuditEventAccountLocked      AuditEvent = "account_locked"
	AuditEventAccountUnlocked    AuditEvent = "account_unlocked"
	AuditEventEmailVerified      AuditEvent = "email_verified"
	AuditEventProfileUpdated     AuditEvent = "profile_updated"
)

type AuditLog struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Event       string     `json:"event" gorm:"not null;size:50;index"`
	IPAddress   string     `json:"ip_address" gorm:"size:45;index"`
	UserAgent   string     `json:"user_agent" gorm:"size:255"`
	Details     string     `json:"details" gorm:"type:text"`
	Success     bool       `json:"success" gorm:"index"`
	RiskLevel   string     `json:"risk_level" gorm:"size:20;index"`
	SessionID   *uuid.UUID `json:"session_id" gorm:"type:uuid;index"`
	Location    string     `json:"location" gorm:"size:255"`
	DeviceInfo  string     `json:"device_info" gorm:"size:255"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

func (a *AuditService) LogEvent(userID *uuid.UUID, event AuditEvent, ipAddress, userAgent, details string, success bool) error {
	riskLevel := a.calculateRiskLevel(event, success)
	
	auditLog := AuditLog{
		UserID:    userID,
		Event:     string(event),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Success:   success,
		RiskLevel: riskLevel,
		Location:  a.getLocationFromIP(ipAddress),
		DeviceInfo: a.extractDeviceInfo(userAgent),
	}
	
	return a.db.Create(&auditLog).Error
}

func (a *AuditService) LogEventWithSession(userID *uuid.UUID, sessionID *uuid.UUID, event AuditEvent, ipAddress, userAgent, details string, success bool) error {
	riskLevel := a.calculateRiskLevel(event, success)
	
	auditLog := AuditLog{
		UserID:    userID,
		SessionID: sessionID,
		Event:     string(event),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Success:   success,
		RiskLevel: riskLevel,
		Location:  a.getLocationFromIP(ipAddress),
		DeviceInfo: a.extractDeviceInfo(userAgent),
	}
	
	return a.db.Create(&auditLog).Error
}

func (a *AuditService) GetUserAuditLogs(userID uuid.UUID, limit int, offset int) ([]AuditLog, error) {
	var logs []AuditLog
	err := a.db.Where("user_id = ?", userID).
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (a *AuditService) GetSecurityEventsAudit(userID uuid.UUID, limit int) ([]AuditLog, error) {
	securityEvents := []string{
		string(AuditEventLoginFailed),
		string(AuditEventMFAFailed),
		string(AuditEventSuspiciousActivity),
		string(AuditEventAccountLocked),
		string(AuditEventSessionRevoked),
	}
	
	var logs []AuditLog
	err := a.db.Where("user_id = ? AND event IN ?", userID, securityEvents).
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (a *AuditService) GetHighRiskEvents(limit int) ([]AuditLog, error) {
	var logs []AuditLog
	err := a.db.Where("risk_level = ?", "high").
		Order("created_at desc").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (a *AuditService) calculateRiskLevel(event AuditEvent, success bool) string {
	if !success {
		switch event {
		case AuditEventLoginFailed, AuditEventMFAFailed:
			return "medium"
		case AuditEventSuspiciousActivity, AuditEventAccountLocked:
			return "high"
		}
	}
	
	switch event {
	case AuditEventLogin, AuditEventLogout, AuditEventEmailVerified:
		return "low"
	case AuditEventPasswordChange, AuditEventMFASetup, AuditEventMFADisable:
		return "medium"
	case AuditEventPasswordReset, AuditEventOAuthLink, AuditEventSessionRevoked:
		return "medium"
	case AuditEventSuspiciousActivity, AuditEventAccountLocked:
		return "high"
	default:
		return "low"
	}
}

func (a *AuditService) getLocationFromIP(ipAddress string) string {
	// In production, use a GeoIP service
	if ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "Local"
	}
	
	// Check if it's a private IP
	ip := net.ParseIP(ipAddress)
	if ip != nil && ip.IsPrivate() {
		return "Private Network"
	}
	
	return "Unknown"
}

func (a *AuditService) extractDeviceInfo(userAgent string) string {
	// Simple device/browser detection
	userAgent = strings.ToLower(userAgent)
	
	browser := "Unknown"
	if strings.Contains(userAgent, "chrome") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "safari") {
		browser = "Safari"
	} else if strings.Contains(userAgent, "edge") {
		browser = "Edge"
	}
	
	os := "Unknown"
	if strings.Contains(userAgent, "windows") {
		os = "Windows"
	} else if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		os = "macOS"
	} else if strings.Contains(userAgent, "linux") {
		os = "Linux"
	} else if strings.Contains(userAgent, "android") {
		os = "Android"
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		os = "iOS"
	}
	
	return fmt.Sprintf("%s on %s", browser, os)
}

// Cleanup expired audit logs (should be run periodically)
func (a *AuditService) CleanupOldAuditLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := a.db.Where("created_at < ?", cutoff).Delete(&AuditLog{})
	return result.Error
}

// Security metrics
type SecurityMetrics struct {
	FailedLogins24h      int64 `json:"failed_logins_24h"`
	SuspiciousActivity24h int64 `json:"suspicious_activity_24h"`
	AccountLockouts24h   int64 `json:"account_lockouts_24h"`
	ActiveSessions       int64 `json:"active_sessions"`
	UnusualLocations24h  int64 `json:"unusual_locations_24h"`
}

func (a *AuditService) GetSecurityMetrics() (*SecurityMetrics, error) {
	metrics := &SecurityMetrics{}
	yesterday := time.Now().Add(-24 * time.Hour)
	
	// Failed logins in last 24h
	a.db.Model(&AuditLog{}).
		Where("event = ? AND success = false AND created_at > ?", AuditEventLoginFailed, yesterday).
		Count(&metrics.FailedLogins24h)
	
	// Suspicious activity in last 24h
	a.db.Model(&AuditLog{}).
		Where("event = ? AND created_at > ?", AuditEventSuspiciousActivity, yesterday).
		Count(&metrics.SuspiciousActivity24h)
	
	// Account lockouts in last 24h
	a.db.Model(&AuditLog{}).
		Where("event = ? AND created_at > ?", AuditEventAccountLocked, yesterday).
		Count(&metrics.AccountLockouts24h)
	
	return metrics, nil
}

// Periodic cleanup function for security service
func (s *SecurityService) RunPeriodicCleanup() {
	// Cleanup rate limiters
	s.loginLimiter.Cleanup()
	s.registrationLimiter.Cleanup()
	s.passwordResetLimiter.Cleanup()
	s.mfaLimiter.Cleanup()
	s.generalLimiter.Cleanup()
}