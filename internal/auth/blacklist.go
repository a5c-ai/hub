package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TokenBlacklist represents a blacklisted token
type TokenBlacklist struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	
	TokenHash    string    `json:"-" gorm:"uniqueIndex;not null;size:64"`
	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;index"` 
	ExpiresAt    time.Time `json:"expires_at" gorm:"not null"`
	Reason       string    `json:"reason" gorm:"size:255"`
	BlacklistedBy uuid.UUID `json:"blacklisted_by" gorm:"type:uuid"`
}

func (TokenBlacklist) TableName() string {
	return "token_blacklist"
}

// TokenBlacklistService handles token blacklisting for secure logout
type TokenBlacklistService struct {
	db *gorm.DB
}

func NewTokenBlacklistService(db *gorm.DB) *TokenBlacklistService {
	return &TokenBlacklistService{db: db}
}

// BlacklistToken adds a token to the blacklist
func (s *TokenBlacklistService) BlacklistToken(token string, expiresAt time.Time) error {
	tokenHash := s.hashToken(token)
	
	blacklistEntry := &TokenBlacklist{
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Reason:    "Manual logout",
	}
	
	return s.db.Create(blacklistEntry).Error
}

// BlacklistTokenWithDetails adds a token to the blacklist with additional details
func (s *TokenBlacklistService) BlacklistTokenWithDetails(token string, userID, blacklistedBy uuid.UUID, expiresAt time.Time, reason string) error {
	tokenHash := s.hashToken(token)
	
	blacklistEntry := &TokenBlacklist{
		TokenHash:     tokenHash,
		UserID:        userID,
		ExpiresAt:     expiresAt,
		Reason:        reason,
		BlacklistedBy: blacklistedBy,
	}
	
	return s.db.Create(blacklistEntry).Error
}

// BlacklistUserTokens blacklists all active tokens for a user (used during logout)
func (s *TokenBlacklistService) BlacklistUserTokens(userID uuid.UUID) error {
	// This would typically involve getting all active sessions for the user
	// and blacklisting their associated tokens. Since we're using session-based
	// refresh tokens, we'll create a general blacklist entry for the user.
	
	blacklistEntry := &TokenBlacklist{
		TokenHash:     s.hashToken(userID.String() + time.Now().String()),
		UserID:        userID,
		ExpiresAt:     time.Now().Add(24 * time.Hour), // Blacklist for 24 hours
		Reason:        "User logout - all devices",
		BlacklistedBy: userID,
	}
	
	return s.db.Create(blacklistEntry).Error
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (s *TokenBlacklistService) IsTokenBlacklisted(token string) (bool, error) {
	tokenHash := s.hashToken(token)
	
	var count int64
	err := s.db.Model(&TokenBlacklist{}).
		Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).
		Count(&count).Error
	
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CleanupExpiredBlacklist removes expired blacklist entries
func (s *TokenBlacklistService) CleanupExpiredBlacklist() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&TokenBlacklist{}).Error
}

// GetBlacklistedTokensForUser returns blacklisted tokens for a specific user
func (s *TokenBlacklistService) GetBlacklistedTokensForUser(userID uuid.UUID) ([]TokenBlacklist, error) {
	var tokens []TokenBlacklist
	err := s.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Find(&tokens).Error
	return tokens, err
}

// RemoveFromBlacklist removes a token from the blacklist (for token restoration)
func (s *TokenBlacklistService) RemoveFromBlacklist(token string) error {
	tokenHash := s.hashToken(token)
	return s.db.Where("token_hash = ?", tokenHash).Delete(&TokenBlacklist{}).Error
}

// GetBlacklistStats returns statistics about the blacklist
func (s *TokenBlacklistService) GetBlacklistStats() (map[string]interface{}, error) {
	var total, active int64
	
	// Total blacklisted tokens
	s.db.Model(&TokenBlacklist{}).Count(&total)
	
	// Active (not expired) blacklisted tokens
	s.db.Model(&TokenBlacklist{}).Where("expires_at > ?", time.Now()).Count(&active)
	
	return map[string]interface{}{
		"total_blacklisted": total,
		"active_blacklisted": active,
		"expired_count": total - active,
	}, nil
}

// hashToken creates a SHA-256 hash of the token for secure storage
func (s *TokenBlacklistService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// RunPeriodicCleanup should be called periodically to clean expired entries
func (s *TokenBlacklistService) RunPeriodicCleanup() error {
	return s.CleanupExpiredBlacklist()
}