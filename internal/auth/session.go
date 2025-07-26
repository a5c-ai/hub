package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:(gen_random_uuid())"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	RefreshToken  string    `json:"-" gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt     time.Time `json:"expires_at" gorm:"not null"`
	IPAddress     string    `json:"ip_address" gorm:"size:45"`
	UserAgent     string    `json:"user_agent" gorm:"size:255"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	LastUsedAt    time.Time `json:"last_used_at"`
	DeviceName    string    `json:"device_name" gorm:"size:255"`
	LocationInfo  string    `json:"location_info" gorm:"size:255"`
	IsRemembered  bool      `json:"is_remembered" gorm:"default:false"`
	SecurityFlags int       `json:"security_flags" gorm:"default:0"`
}

func (Session) TableName() string {
	return "sessions"
}

type SessionService struct {
	db     *gorm.DB
	config *SessionConfig
}

type SessionConfig struct {
	MaxSessions          int           `json:"max_sessions"`
	DefaultExpiration    time.Duration `json:"default_expiration"`
	RememberMeExpiration time.Duration `json:"remember_me_expiration"`
	IdleTimeout          time.Duration `json:"idle_timeout"`
	RequireSecureHeaders bool          `json:"require_secure_headers"`
	EnableGeoTracking    bool          `json:"enable_geo_tracking"`
	EnableDeviceTracking bool          `json:"enable_device_tracking"`
	AutoCleanupInterval  time.Duration `json:"auto_cleanup_interval"`
}

func NewSessionService(db *gorm.DB) *SessionService {
	defaultConfig := &SessionConfig{
		MaxSessions:          5,
		DefaultExpiration:    30 * 24 * time.Hour, // 30 days
		RememberMeExpiration: 90 * 24 * time.Hour, // 90 days
		IdleTimeout:          24 * time.Hour,      // 24 hours
		RequireSecureHeaders: true,
		EnableGeoTracking:    false,
		EnableDeviceTracking: true,
		AutoCleanupInterval:  1 * time.Hour,
	}

	return &SessionService{
		db:     db,
		config: defaultConfig,
	}
}

func NewSessionServiceWithConfig(db *gorm.DB, config *SessionConfig) *SessionService {
	return &SessionService{
		db:     db,
		config: config,
	}
}

func (s *SessionService) CreateSession(userID uuid.UUID, ipAddress, userAgent string, rememberMe bool) (*Session, error) {
	// Generate secure refresh token
	refreshToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Set expiration time based on remember me preference
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(s.config.RememberMeExpiration)
	} else {
		expiresAt = time.Now().Add(s.config.DefaultExpiration)
	}

	// Extract device name from user agent
	deviceName := s.extractDeviceName(userAgent)

	// Create session
	session := &Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		DeviceName:   deviceName,
		IsActive:     true,
		IsRemembered: rememberMe,
		LastUsedAt:   time.Now(),
	}

	// Get location info if enabled
	if s.config.EnableGeoTracking {
		session.LocationInfo = s.getLocationInfo(ipAddress)
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Limit concurrent sessions for user
	if err := s.LimitUserSessions(userID, s.config.MaxSessions); err != nil {
		// Log error but don't fail session creation
		fmt.Printf("Failed to limit user sessions: %v\n", err)
	}

	return session, nil
}

func (s *SessionService) ValidateRefreshToken(refreshToken string) (*Session, error) {
	var session Session
	err := s.db.Where("refresh_token = ? AND is_active = true AND expires_at > ?",
		refreshToken, time.Now()).First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Update last used time
	session.LastUsedAt = time.Now()
	s.db.Save(&session)

	return &session, nil
}

func (s *SessionService) RefreshSession(refreshToken string) (*Session, error) {
	// Validate current token
	session, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Update session with new token and extended expiration based on configuration
	session.RefreshToken = newRefreshToken
	if session.IsRemembered {
		session.ExpiresAt = time.Now().Add(s.config.RememberMeExpiration)
	} else {
		session.ExpiresAt = time.Now().Add(s.config.DefaultExpiration)
	}
	session.LastUsedAt = time.Now()

	if err := s.db.Save(session).Error; err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

func (s *SessionService) RevokeSession(refreshToken string) error {
	return s.db.Model(&Session{}).
		Where("refresh_token = ?", refreshToken).
		Update("is_active", false).Error
}

func (s *SessionService) RevokeUserSessions(userID uuid.UUID) error {
	return s.db.Model(&Session{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error
}

func (s *SessionService) GetUserSessions(userID uuid.UUID) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("user_id = ? AND is_active = true AND expires_at > ?",
		userID, time.Now()).Find(&sessions).Error
	return sessions, err
}

func (s *SessionService) CleanupExpiredSessions() error {
	// Delete expired sessions older than 7 days
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	result := s.db.Where("expires_at < ? OR (is_active = false AND updated_at < ?)",
		time.Now(), cutoff).Delete(&Session{})
	return result.Error
}

func (s *SessionService) LimitUserSessions(userID uuid.UUID, maxSessions int) error {
	// Get user sessions ordered by last used (oldest first)
	var sessions []Session
	err := s.db.Where("user_id = ? AND is_active = true AND expires_at > ?",
		userID, time.Now()).
		Order("last_used_at asc").
		Find(&sessions).Error

	if err != nil {
		return err
	}

	// If user has more sessions than allowed, revoke the oldest ones
	if len(sessions) > maxSessions {
		sessionsToRevoke := sessions[:len(sessions)-maxSessions]
		for _, session := range sessionsToRevoke {
			s.RevokeSession(session.RefreshToken)
		}
	}

	return nil
}

func (s *SessionService) generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Session management for rate limiting and security
type SessionStats struct {
	TotalSessions  int64     `json:"total_sessions"`
	ActiveSessions int64     `json:"active_sessions"`
	LastCleanup    time.Time `json:"last_cleanup"`
}

func (s *SessionService) GetSessionStats() (*SessionStats, error) {
	var total, active int64

	// Get total sessions
	s.db.Model(&Session{}).Count(&total)

	// Get active sessions
	s.db.Model(&Session{}).Where("is_active = true AND expires_at > ?", time.Now()).Count(&active)

	return &SessionStats{
		TotalSessions:  total,
		ActiveSessions: active,
		LastCleanup:    time.Now(),
	}, nil
}

// Enhanced session validation with idle timeout
func (s *SessionService) ValidateSessionWithIdleCheck(refreshToken string) (*Session, error) {
	// Validate token without updating last used time
	var session Session
	err := s.db.Where("refresh_token = ? AND is_active = true AND expires_at > ?", refreshToken, time.Now()).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check idle timeout
	if time.Since(session.LastUsedAt) > s.config.IdleTimeout {
		// Session is idle, revoke it
		s.RevokeSession(refreshToken)
		return nil, errors.New("session expired due to inactivity")
	}

	// Update last used time
	session.LastUsedAt = time.Now()
	s.db.Save(&session)

	return &session, nil
}

// Detect suspicious activity
func (s *SessionService) DetectSuspiciousActivity(userID uuid.UUID, ipAddress string) (bool, error) {
	// Get recent sessions for user
	var sessions []Session
	err := s.db.Where("user_id = ? AND is_active = true AND created_at > ?",
		userID, time.Now().Add(-24*time.Hour)).Find(&sessions).Error
	if err != nil {
		return false, err
	}

	// Check for multiple different IP addresses in short time
	ipMap := make(map[string]bool)
	for _, session := range sessions {
		ipMap[session.IPAddress] = true
	}

	// If more than 3 different IPs in 24 hours, flag as suspicious
	if len(ipMap) > 3 {
		return true, nil
	}

	return false, nil
}

// Force logout from all devices
func (s *SessionService) ForceLogoutAllDevices(userID uuid.UUID) error {
	return s.RevokeUserSessions(userID)
}

// Get detailed session information for security dashboard
func (s *SessionService) GetDetailedUserSessions(userID uuid.UUID) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("user_id = ? AND is_active = true AND expires_at > ?",
		userID, time.Now()).
		Order("last_used_at desc").
		Find(&sessions).Error
	return sessions, err
}

// Update session activity (called on each API request)
func (s *SessionService) UpdateSessionActivity(refreshToken, ipAddress string) error {
	return s.db.Model(&Session{}).
		Where("refresh_token = ? AND is_active = true", refreshToken).
		Updates(map[string]interface{}{
			"last_used_at": time.Now(),
			"ip_address":   ipAddress,
		}).Error
}

// Helper methods
func (s *SessionService) extractDeviceName(userAgent string) string {
	// Simple device detection based on user agent
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") {
		return "Mobile Device"
	}
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "iOS Device"
	}
	if strings.Contains(userAgent, "windows") {
		return "Windows Computer"
	}
	if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
		return "Mac Computer"
	}
	if strings.Contains(userAgent, "linux") {
		return "Linux Computer"
	}

	return "Unknown Device"
}

func (s *SessionService) getLocationInfo(ipAddress string) string {
	// In production, you would use a GeoIP service
	// For now, return a placeholder
	if ipAddress == "127.0.0.1" || ipAddress == "::1" {
		return "Local"
	}

	// This is where you'd integrate with a GeoIP service like MaxMind
	return "Unknown Location"
}

// Automatic session cleanup (should be run periodically)
func (s *SessionService) RunPeriodicCleanup() error {
	// Clean up expired sessions
	if err := s.CleanupExpiredSessions(); err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	// Clean up idle sessions
	idleCutoff := time.Now().Add(-s.config.IdleTimeout)
	err := s.db.Model(&Session{}).
		Where("last_used_at < ? AND is_active = true", idleCutoff).
		Update("is_active", false).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup idle sessions: %w", err)
	}

	return nil
}

// Session security flags
const (
	SessionFlagNormal         = 0
	SessionFlagSuspicious     = 1 << 0
	SessionFlagCompromised    = 1 << 1
	SessionFlagLocationChange = 1 << 2
	SessionFlagDeviceChange   = 1 << 3
)

func (s *SessionService) FlagSession(sessionID uuid.UUID, flag int) error {
	return s.db.Model(&Session{}).
		Where("id = ?", sessionID).
		Update("security_flags", gorm.Expr("security_flags | ?", flag)).Error
}

func (s *SessionService) GetFlaggedSessions(userID uuid.UUID) ([]Session, error) {
	var sessions []Session
	err := s.db.Where("user_id = ? AND security_flags > 0 AND is_active = true", userID).
		Find(&sessions).Error
	return sessions, err
}
