package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	RefreshToken string    `json:"-" gorm:"not null;uniqueIndex;size:255"`
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:255"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastUsedAt   time.Time `json:"last_used_at"`
}

func (Session) TableName() string {
	return "sessions"
}

type SessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

func (s *SessionService) CreateSession(userID uuid.UUID, ipAddress, userAgent string) (*Session, error) {
	// Generate secure refresh token
	refreshToken, err := s.generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Set expiration time (30 days from now)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	// Create session
	session := &Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		IsActive:     true,
		LastUsedAt:   time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
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

	// Update session with new token and extended expiration
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)
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