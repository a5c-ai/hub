package auth

import (
	"fmt"

	"github.com/a5c-ai/hub/internal/models"
	"gorm.io/gorm"
)

// MigrateAuthTables runs all authentication-related database migrations
func MigrateAuthTables(db *gorm.DB) error {
	// List of all models that need to be migrated
	models := []interface{}{
		&models.User{},
		&Session{},
		&BackupCode{},
		&WebAuthnCredential{},
		&SMSVerificationCode{},
		&EmailVerificationToken{},
		&PasswordResetToken{},
		&OAuthAccount{},
		&OAuthState{},
		&SecurityEvent{},
		&LoginAttempt{},
		&AuditLog{},
		&AccountLockout{},
	}

	// Run auto-migration for all models
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
	}

	// Create indexes for better performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email_active ON users(email, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_username_active ON users(username, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login_at)",
		
		// Session indexes
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_expires_active ON sessions(expires_at, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_last_used ON sessions(last_used_at)",
		"CREATE INDEX IF NOT EXISTS idx_sessions_ip_created ON sessions(ip_address, created_at)",
		
		// MFA indexes
		"CREATE INDEX IF NOT EXISTS idx_backup_codes_user_used ON backup_codes(user_id, used)",
		"CREATE INDEX IF NOT EXISTS idx_webauthn_credentials_user ON webauthn_credentials(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_sms_codes_user_expires ON sms_verification_codes(user_id, expires_at, used)",
		
		// Email verification indexes
		"CREATE INDEX IF NOT EXISTS idx_email_tokens_expires_used ON email_verification_tokens(expires_at, used)",
		"CREATE INDEX IF NOT EXISTS idx_password_reset_expires_used ON password_reset_tokens(expires_at, used)",
		
		// OAuth indexes
		"CREATE INDEX IF NOT EXISTS idx_oauth_accounts_user_provider ON oauth_accounts(user_id, provider)",
		"CREATE INDEX IF NOT EXISTS idx_oauth_accounts_provider_id ON oauth_accounts(provider, provider_id)",
		"CREATE INDEX IF NOT EXISTS idx_oauth_states_expires_used ON oauth_states(expires_at, used)",
		
		// Security indexes
		"CREATE INDEX IF NOT EXISTS idx_security_events_user_created ON security_events(user_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_security_events_severity_created ON security_events(severity, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_security_events_ip_created ON security_events(ip_address, created_at)",
		
		// Login attempt indexes
		"CREATE INDEX IF NOT EXISTS idx_login_attempts_email_created ON login_attempts(email, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_created ON login_attempts(ip_address, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_login_attempts_success_created ON login_attempts(success, created_at)",
		
		// Audit log indexes
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_event_created ON audit_logs(event, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_risk_created ON audit_logs(risk_level, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_created ON audit_logs(ip_address, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_session_created ON audit_logs(session_id, created_at)",
		
		// Account lockout indexes
		"CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_active ON account_lockouts(user_id, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_account_lockouts_ip_active ON account_lockouts(ip_address, is_active)",
		"CREATE INDEX IF NOT EXISTS idx_account_lockouts_locked_until ON account_lockouts(locked_until)",
	}

	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			// Log the error but don't fail migration if index already exists
			fmt.Printf("Warning: Failed to create index: %v\n", err)
		}
	}

	return nil
}

// CleanupExpiredData removes expired data from authentication tables
func CleanupExpiredData(db *gorm.DB) error {
	// Clean up expired email verification tokens (older than 30 days)
	if err := db.Where("expires_at < ? OR (used = true AND updated_at < ?)", 
		getTimeAgo(30*24), getTimeAgo(7*24)).Delete(&EmailVerificationToken{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup email verification tokens: %w", err)
	}

	// Clean up expired password reset tokens (older than 7 days)
	if err := db.Where("expires_at < ? OR (used = true AND updated_at < ?)", 
		getTimeAgo(24), getTimeAgo(24)).Delete(&PasswordResetToken{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup password reset tokens: %w", err)
	}

	// Clean up expired SMS verification codes (older than 1 day)
	if err := db.Where("expires_at < ? OR (used = true AND updated_at < ?)", 
		getTimeAgo(1), getTimeAgo(1)).Delete(&SMSVerificationCode{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup SMS verification codes: %w", err)
	}

	// Clean up expired OAuth states (older than 1 hour)
	if err := db.Where("expires_at < ? OR (used = true AND updated_at < ?)", 
		getTimeAgo(0.04), getTimeAgo(0.04)).Delete(&OAuthState{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup OAuth states: %w", err)
	}

	// Clean up old sessions (expired or inactive for more than 7 days)
	if err := db.Where("expires_at < ? OR (is_active = false AND updated_at < ?)", 
		getTimeAgo(0), getTimeAgo(7*24)).Delete(&Session{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup sessions: %w", err)
	}

	// Clean up old security events (older than 90 days)
	if err := db.Where("created_at < ?", getTimeAgo(90*24)).Delete(&SecurityEvent{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup security events: %w", err)
	}

	// Clean up old login attempts (older than 30 days)
	if err := db.Where("created_at < ?", getTimeAgo(30*24)).Delete(&LoginAttempt{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup login attempts: %w", err)
	}

	// Clean up old audit logs (older than 365 days)
	if err := db.Where("created_at < ?", getTimeAgo(365*24)).Delete(&AuditLog{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup audit logs: %w", err)
	}

	// Clean up resolved account lockouts (older than 30 days)
	if err := db.Where("is_active = false AND updated_at < ?", getTimeAgo(30*24)).Delete(&AccountLockout{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup account lockouts: %w", err)
	}

	return nil
}

// Helper function to get time ago in hours
func getTimeAgo(hours float64) interface{} {
	return fmt.Sprintf("datetime('now', '-%v hours')", hours)
}

// GetDatabaseStats returns statistics about authentication tables
type DatabaseStats struct {
	TotalUsers           int64 `json:"total_users"`
	ActiveUsers          int64 `json:"active_users"`
	VerifiedUsers        int64 `json:"verified_users"`
	MFAEnabledUsers      int64 `json:"mfa_enabled_users"`
	ActiveSessions       int64 `json:"active_sessions"`
	TotalSessions        int64 `json:"total_sessions"`
	OAuthAccounts        int64 `json:"oauth_accounts"`
	RecentLogins24h      int64 `json:"recent_logins_24h"`
	FailedLogins24h      int64 `json:"failed_logins_24h"`
	SecurityEvents24h    int64 `json:"security_events_24h"`
	PendingVerifications int64 `json:"pending_verifications"`
}

func GetDatabaseStats(db *gorm.DB) (*DatabaseStats, error) {
	stats := &DatabaseStats{}
	
	// Total users
	db.Model(&models.User{}).Count(&stats.TotalUsers)
	
	// Active users
	db.Model(&models.User{}).Where("is_active = true").Count(&stats.ActiveUsers)
	
	// Verified users
	db.Model(&models.User{}).Where("email_verified = true").Count(&stats.VerifiedUsers)
	
	// MFA enabled users
	db.Model(&models.User{}).Where("two_factor_enabled = true").Count(&stats.MFAEnabledUsers)
	
	// Sessions
	db.Model(&Session{}).Count(&stats.TotalSessions)
	db.Model(&Session{}).Where("is_active = true AND expires_at > datetime('now')").Count(&stats.ActiveSessions)
	
	// OAuth accounts
	db.Model(&OAuthAccount{}).Count(&stats.OAuthAccounts)
	
	// Recent activity (last 24 hours)
	db.Model(&AuditLog{}).Where("event = 'login' AND success = true AND created_at > datetime('now', '-24 hours')").Count(&stats.RecentLogins24h)
	db.Model(&AuditLog{}).Where("event = 'login_failed' AND created_at > datetime('now', '-24 hours')").Count(&stats.FailedLogins24h)
	db.Model(&SecurityEvent{}).Where("created_at > datetime('now', '-24 hours')").Count(&stats.SecurityEvents24h)
	
	// Pending verifications
	db.Model(&EmailVerificationToken{}).Where("used = false AND expires_at > datetime('now')").Count(&stats.PendingVerifications)
	
	return stats, nil
}

// ValidateDatabaseIntegrity checks for common database integrity issues
func ValidateDatabaseIntegrity(db *gorm.DB) []string {
	var issues []string
	
	// Check for users without valid sessions but marked as recently active
	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM users u 
		WHERE u.last_login_at > datetime('now', '-1 hours') 
		AND NOT EXISTS (
			SELECT 1 FROM sessions s 
			WHERE s.user_id = u.id 
			AND s.is_active = true 
			AND s.expires_at > datetime('now')
		)
	`).Scan(&count)
	if count > 0 {
		issues = append(issues, fmt.Sprintf("%d users marked as recently active but have no active sessions", count))
	}
	
	// Check for expired tokens that haven't been cleaned up
	db.Model(&EmailVerificationToken{}).Where("expires_at < datetime('now', '-30 days')").Count(&count)
	if count > 0 {
		issues = append(issues, fmt.Sprintf("%d expired email verification tokens need cleanup", count))
	}
	
	// Check for orphaned backup codes
	db.Raw(`
		SELECT COUNT(*) FROM backup_codes b 
		WHERE NOT EXISTS (
			SELECT 1 FROM users u 
			WHERE u.id = b.user_id
		)
	`).Scan(&count)
	if count > 0 {
		issues = append(issues, fmt.Sprintf("%d orphaned backup codes found", count))
	}
	
	// Check for users with MFA enabled but no backup codes
	db.Raw(`
		SELECT COUNT(*) FROM users u 
		WHERE u.two_factor_enabled = true 
		AND NOT EXISTS (
			SELECT 1 FROM backup_codes b 
			WHERE b.user_id = u.id AND b.used = false
		)
	`).Scan(&count)
	if count > 0 {
		issues = append(issues, fmt.Sprintf("%d users with MFA enabled but no backup codes", count))
	}
	
	return issues
}

// RepairDatabaseIntegrity attempts to fix common database integrity issues
func RepairDatabaseIntegrity(db *gorm.DB) error {
	// Remove orphaned records
	if err := db.Exec(`
		DELETE FROM backup_codes 
		WHERE user_id NOT IN (SELECT id FROM users)
	`).Error; err != nil {
		return fmt.Errorf("failed to remove orphaned backup codes: %w", err)
	}
	
	if err := db.Exec(`
		DELETE FROM sessions 
		WHERE user_id NOT IN (SELECT id FROM users)
	`).Error; err != nil {
		return fmt.Errorf("failed to remove orphaned sessions: %w", err)
	}
	
	if err := db.Exec(`
		DELETE FROM oauth_accounts 
		WHERE user_id NOT IN (SELECT id FROM users)
	`).Error; err != nil {
		return fmt.Errorf("failed to remove orphaned OAuth accounts: %w", err)
	}
	
	// Deactivate expired sessions
	if err := db.Model(&Session{}).
		Where("expires_at < datetime('now') AND is_active = true").
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate expired sessions: %w", err)
	}
	
	return nil
}