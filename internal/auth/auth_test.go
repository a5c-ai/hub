package auth

import (
	"context"
	"testing"
	"time"

	"database/sql"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testSQLiteDriver is a custom SQLite driver name used to register a SQLite3 driver with gen_random_uuid() support
const testSQLiteDriver = "sqlite3_gen_random_uuid"

func init() {
	// Register custom SQLite driver with gen_random_uuid() support for tests
	sql.Register(testSQLiteDriver, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			conn.RegisterFunc("gen_random_uuid", func() string {
				return uuid.New().String()
			}, true)
			return nil
		},
	})
}

func setupTestDB(t *testing.T) *gorm.DB {
	// Open in-memory SQLite DB using the custom driver supporting gen_random_uuid()
	dialector := sqlite.Open(":memory:")
	if dr, ok := dialector.(*sqlite.Dialector); ok {
		dr.DriverName = testSQLiteDriver
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Migrate tables
	err = db.AutoMigrate(
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
	)
	require.NoError(t, err)

	return db
}

func setupTestServices(t *testing.T) (AuthService, *gorm.DB, *config.Config) {
	db := setupTestDB(t)
	cfg := &config.Config{
		JWT: config.JWT{
			Secret:         "test-secret",
			ExpirationHour: 24,
		},
		SMTP: config.SMTP{
			Host:     "",
			Port:     "587",
			Username: "",
			Password: "",
			From:     "test@example.com",
			UseTLS:   false,
		},
	}

	jwtManager := NewJWTManager(cfg.JWT)
	authService := NewAuthService(db, jwtManager, cfg)

	return authService, db, cfg
}

func TestUserRegistration(t *testing.T) {
	authService, _, _ := setupTestServices(t)

	tests := []struct {
		name        string
		request     RegisterRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "SecurePassword123!",
				FullName: "Test User",
			},
			expectError: false,
		},
		{
			name: "duplicate email",
			request: RegisterRequest{
				Username: "testuser2",
				Email:    "test@example.com",
				Password: "SecurePassword123!",
				FullName: "Test User 2",
			},
			expectError: true,
		},
		{
			name: "duplicate username",
			request: RegisterRequest{
				Username: "testuser",
				Email:    "test2@example.com",
				Password: "SecurePassword123!",
				FullName: "Test User 3",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := authService.Register(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.Username, user.Username)
				assert.Equal(t, tt.request.Email, user.Email)
				assert.Equal(t, tt.request.FullName, user.FullName)
				assert.True(t, user.IsActive)
				assert.False(t, user.IsAdmin)
				assert.Empty(t, user.PasswordHash) // Should be removed in response
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	authService, _, _ := setupTestServices(t)

	// First register a user
	user, err := authService.Register(context.Background(), RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "SecurePassword123!",
		FullName: "Test User",
	})
	require.NoError(t, err)
	require.NotNil(t, user)

	tests := []struct {
		name        string
		request     LoginRequest
		expectError bool
	}{
		{
			name: "valid login with email",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "SecurePassword123!",
			},
			expectError: false,
		},
		{
			name: "invalid password",
			request: LoginRequest{
				Email:    "test@example.com",
				Password: "WrongPassword",
			},
			expectError: true,
		},
		{
			name: "invalid email",
			request: LoginRequest{
				Email:    "wrong@example.com",
				Password: "SecurePassword123!",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := authService.Login(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, user.ID, response.User.ID)
				assert.Empty(t, response.User.PasswordHash)
			}
		})
	}
}

func TestMFASetup(t *testing.T) {
	_, db, _ := setupTestServices(t)
	mfaService := NewMFAService(db)

	userID := uuid.New()

	// Test TOTP setup
	t.Run("TOTP setup", func(t *testing.T) {
		response, err := mfaService.SetupTOTP(userID, "Test App", "test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Secret)
		assert.NotEmpty(t, response.QRCodeURL)
		assert.Len(t, response.BackupCodes, 10)

		// Verify backup codes were stored
		var codes []BackupCode
		err = db.Where("user_id = ?", userID).Find(&codes).Error
		assert.NoError(t, err)
		assert.Len(t, codes, 10)
	})

	// Test backup code usage
	t.Run("backup code usage", func(t *testing.T) {
		// Get a backup code
		var code BackupCode
		err := db.Where("user_id = ? AND used = false", userID).First(&code).Error
		require.NoError(t, err)

		// Use the backup code
		valid, err := mfaService.useBackupCode(userID, code.Code)
		assert.NoError(t, err)
		assert.True(t, valid)

		// Verify code is marked as used
		err = db.First(&code).Error
		assert.NoError(t, err)
		assert.True(t, code.Used)
		assert.NotNil(t, code.UsedAt)

		// Try to use the same code again
		valid, err = mfaService.useBackupCode(userID, code.Code)
		assert.NoError(t, err)
		assert.False(t, valid)
	})
}

func TestSessionManagement(t *testing.T) {
	_, db, _ := setupTestServices(t)
	sessionService := NewSessionService(db)

	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// Test session creation
	t.Run("create session", func(t *testing.T) {
		session, err := sessionService.CreateSession(userID, ipAddress, userAgent, false)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		assert.Equal(t, userID, session.UserID)
		assert.Equal(t, ipAddress, session.IPAddress)
		assert.True(t, session.IsActive)
		assert.False(t, session.IsRemembered)
		assert.NotEmpty(t, session.RefreshToken)
	})

	// Test session validation
	t.Run("validate session", func(t *testing.T) {
		// Create a session first
		session, err := sessionService.CreateSession(userID, ipAddress, userAgent, true)
		require.NoError(t, err)

		// Validate the session
		validatedSession, err := sessionService.ValidateRefreshToken(session.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, validatedSession)
		assert.Equal(t, session.ID, validatedSession.ID)
	})

	// Test session refresh
	t.Run("refresh session", func(t *testing.T) {
		// Create a session first
		originalSession, err := sessionService.CreateSession(userID, ipAddress, userAgent, false)
		require.NoError(t, err)

		// Refresh the session
		refreshedSession, err := sessionService.RefreshSession(originalSession.RefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, refreshedSession)
		assert.Equal(t, originalSession.ID, refreshedSession.ID)
		assert.NotEqual(t, originalSession.RefreshToken, refreshedSession.RefreshToken)
	})

	// Test session revocation
	t.Run("revoke session", func(t *testing.T) {
		// Create a session first
		session, err := sessionService.CreateSession(userID, ipAddress, userAgent, false)
		require.NoError(t, err)

		// Revoke the session
		err = sessionService.RevokeSession(session.RefreshToken)
		assert.NoError(t, err)

		// Try to validate the revoked session
		_, err = sessionService.ValidateRefreshToken(session.RefreshToken)
		assert.Error(t, err)
	})

	// Test expired session
	t.Run("expired session", func(t *testing.T) {
		// Create a session first
		session, err := sessionService.CreateSession(userID, ipAddress, userAgent, false)
		require.NoError(t, err)

		// Expire the session
		session.ExpiresAt = time.Now().Add(-time.Hour)
		require.NoError(t, db.Save(session).Error)
		_, err = sessionService.ValidateRefreshToken(session.RefreshToken)
		assert.Error(t, err)
	})

	// Test idle timeout expiration
	t.Run("idle timeout", func(t *testing.T) {
		// Create a session first
		session, err := sessionService.CreateSession(userID, ipAddress, userAgent, false)
		require.NoError(t, err)

		// Simulate idle timeout
		session.LastUsedAt = time.Now().Add(-sessionService.config.IdleTimeout - time.Minute)
		require.NoError(t, db.Save(session).Error)
		_, err = sessionService.ValidateSessionWithIdleCheck(session.RefreshToken)
		assert.Error(t, err)
	})
}

func TestSecurityFeatures(t *testing.T) {
	_, db, _ := setupTestServices(t)
	securityService := NewSecurityService(db)
	auditService := NewAuditService(db)

	userID := uuid.New()
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// Test rate limiting
	t.Run("rate limiting", func(t *testing.T) {
		// First few requests should be allowed
		for i := 0; i < 5; i++ {
			allowed := securityService.CheckLoginRateLimit(ipAddress)
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}

		// 6th request should be blocked
		allowed := securityService.CheckLoginRateLimit(ipAddress)
		assert.False(t, allowed, "6th request should be blocked")
	})

	// Test audit logging
	t.Run("audit logging", func(t *testing.T) {
		// Log an event
		err := auditService.LogEvent(&userID, AuditEventLogin, ipAddress, userAgent, "Test login", true)
		assert.NoError(t, err)

		// Retrieve audit logs
		logs, err := auditService.GetUserAuditLogs(userID, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, string(AuditEventLogin), logs[0].Event)
		assert.Equal(t, userID, *logs[0].UserID)
		assert.True(t, logs[0].Success)
	})

	// Test password strength validation
	t.Run("password strength", func(t *testing.T) {
		tests := []struct {
			password string
			valid    bool
		}{
			{"weak", false},
			{"StrongPassword123!", true},
			{"NoSpecial123", false},
			{"noupppercase123!", false},
			{"NOLOWERCASE123!", false},
			{"NoDigits!", false},
			{"Short1!", false},
		}

		for _, tt := range tests {
			issues := securityService.ValidatePasswordStrength(tt.password)
			if tt.valid {
				assert.Empty(t, issues, "Password %q should be valid", tt.password)
			} else {
				assert.NotEmpty(t, issues, "Password %q should be invalid", tt.password)
			}
		}
	})
}

func TestEmailVerification(t *testing.T) {
	_, db, cfg := setupTestServices(t)
	emailService := NewSMTPEmailService(cfg)
	verificationService := NewEmailVerificationService(db, emailService)

	userID := uuid.New()

	// Test token creation
	t.Run("create verification token", func(t *testing.T) {
		token, err := verificationService.CreateVerificationToken(userID)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, userID, token.UserID)
		assert.NotEmpty(t, token.Token)
		assert.False(t, token.Used)
		assert.True(t, token.ExpiresAt.After(time.Now()))
	})

	// Test email verification
	t.Run("verify email", func(t *testing.T) {
		// Create a token first
		token, err := verificationService.CreateVerificationToken(userID)
		require.NoError(t, err)

		// Create a user in the database
		user := models.User{
			ID:            userID,
			Username:      "testuser",
			Email:         "test@example.com",
			EmailVerified: false,
		}
		err = db.Create(&user).Error
		require.NoError(t, err)

		// Verify the email
		err = verificationService.VerifyEmail(token.Token)
		assert.NoError(t, err)

		// Check that user is marked as verified
		err = db.First(&user).Error
		assert.NoError(t, err)
		assert.True(t, user.EmailVerified)

		// Check that token is marked as used
		err = db.First(token).Error
		assert.NoError(t, err)
		assert.True(t, token.Used)
	})
}

func TestPasswordReset(t *testing.T) {
	_, db, _ := setupTestServices(t)
	passwordResetService := NewPasswordResetService(db)

	userID := uuid.New()

	// Test token creation
	t.Run("create reset token", func(t *testing.T) {
		token, err := passwordResetService.CreateResetToken(userID)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, userID, token.UserID)
		assert.NotEmpty(t, token.Token)
		assert.False(t, token.Used)
		assert.True(t, token.ExpiresAt.After(time.Now()))
	})

	// Test password reset
	t.Run("reset password", func(t *testing.T) {
		// Create a user first
		user := models.User{
			ID:           userID,
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "old-hashed-password",
		}
		err := db.Create(&user).Error
		require.NoError(t, err)

		// Create a reset token
		token, err := passwordResetService.CreateResetToken(userID)
		require.NoError(t, err)

		// Reset the password
		newPassword := "NewSecurePassword123!"
		err = passwordResetService.UseResetToken(token.Token, newPassword)
		assert.NoError(t, err)

		// Check that password was changed
		err = db.First(&user).Error
		assert.NoError(t, err)
		assert.NotEqual(t, "old-hashed-password", user.PasswordHash)

		// Check that token is marked as used
		err = db.First(token).Error
		assert.NoError(t, err)
		assert.True(t, token.Used)
	})
}

func TestOAuthFlow(t *testing.T) {
	authService, db, cfg := setupTestServices(t)
	jwtManager := NewJWTManager(cfg.JWT)
	oauthService := NewOAuthService(db, jwtManager, cfg, authService)

	// Test state generation and validation
	t.Run("oauth state management", func(t *testing.T) {
		state, err := oauthService.GenerateState()
		assert.NoError(t, err)
		assert.NotEmpty(t, state)

		// Store the state
		err = oauthService.StoreState(state, "github")
		assert.NoError(t, err)

		// Validate the state
		err = oauthService.ValidateState(state, "github")
		assert.NoError(t, err)

		// Try to validate again (should fail as it's marked as used)
		err = oauthService.ValidateState(state, "github")
		assert.Error(t, err)
	})

	// Test user creation from OAuth
	t.Run("create user from oauth", func(t *testing.T) {
		userInfo := &OAuthUserInfo{
			ID:       "123456",
			Username: "testuser",
			Email:    "test@example.com",
			Name:     "Test User",
			Avatar:   "https://example.com/avatar.jpg",
		}

		user, err := oauthService.findOrCreateOAuthUser(userInfo, "github")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userInfo.Email, user.Email)
		assert.Equal(t, userInfo.Name, user.FullName)
		assert.True(t, user.EmailVerified) // OAuth emails are considered verified
		assert.Empty(t, user.PasswordHash) // OAuth users don't have passwords
	})
}

// Benchmark tests
func BenchmarkPasswordHashing(b *testing.B) {
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bcryptHashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTokenGeneration(b *testing.B) {
	cfg := &config.Config{
		JWT: config.JWT{
			Secret:         "test-secret",
			ExpirationHour: 24,
		},
	}
	jwtManager := NewJWTManager(cfg.JWT)

	user := &models.User{
		ID:       uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtManager.GenerateToken(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for password hashing (would be in the actual bcrypt utility)
func bcryptHashPassword(password string) (string, error) {
	// This would use golang.org/x/crypto/bcrypt in the real implementation
	// For testing purposes, we'll just return a mock hash
	return "hashed_" + password, nil
}
