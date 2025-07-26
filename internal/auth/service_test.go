package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/a5c-ai/hub/internal/models"
)

// TestGenerateAuthSecureToken verifies that the generated tokens are non-empty, of expected length, and unique.
func TestGenerateAuthSecureToken(t *testing.T) {
	token1 := generateAuthSecureToken()
	token2 := generateAuthSecureToken()
	assert.NotEmpty(t, token1)
	assert.Len(t, token1, 64)
	assert.NotEqual(t, token1, token2)
}

// TestInitiatePasswordReset verifies that InitiatePasswordReset creates a reset token for existing users.
func TestInitiatePasswordReset(t *testing.T) {
	authIface, db, _ := setupTestServices(t)
	authService, ok := authIface.(*authService)
	require.True(t, ok)

	// Create a test user
	userID := uuid.New()
	user := models.User{
		ID:           userID,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "old-hash",
		IsActive:     true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Initiate password reset
	err := authService.InitiatePasswordReset(user.Email)
	assert.NoError(t, err)

	// Verify reset token stored
	var tokens []PasswordResetToken
	err = db.Where("user_id = ?", userID).Find(&tokens).Error
	assert.NoError(t, err)
	assert.Len(t, tokens, 1)
}

// TestInitiatePasswordResetNoUser verifies no error when the user does not exist.
func TestInitiatePasswordResetNoUser(t *testing.T) {
	authIface, _, _ := setupTestServices(t)
	authService, ok := authIface.(*authService)
	require.True(t, ok)

	err := authService.InitiatePasswordReset("nonexistent@example.com")
	assert.NoError(t, err)
}

// TestExtractTokenFromURL verifies that extractTokenFromURL parses a token from an HTML link body.
func TestExtractTokenFromURL(t *testing.T) {
	body := `<a href="http://example.com/reset-password?token=abcd1234">`
	token := extractTokenFromURL(body)
	assert.Equal(t, "abcd1234", token)
}
