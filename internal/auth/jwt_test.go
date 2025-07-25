package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
)

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	testUserID := uuid.New()
	user := &models.User{
		ID:       testUserID,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	token, err := jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.IsAdmin != user.IsAdmin {
		t.Errorf("Expected admin status %t, got %t", user.IsAdmin, claims.IsAdmin)
	}
}

func TestJWTManager_ValidateInvalidToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	_, err := jwtManager.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestJWTManager_ValidateExpiredToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: -1, // Already expired
	}
	jwtManager := NewJWTManager(cfg)

	testUserID := uuid.New()
	user := &models.User{
		ID:       testUserID,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	token, err := jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(time.Millisecond * 100) // Wait a bit to ensure expiration

	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestJWTManager_ValidateEmptyToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	_, err := jwtManager.ValidateToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestJWTManager_ValidateMalformedToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	_, err := jwtManager.ValidateToken("not.a.jwt")
	if err == nil {
		t.Error("Expected error for malformed token, got nil")
	}
}

func TestJWTManager_ValidateTokenWithWrongSecret(t *testing.T) {
	cfg1 := config.JWT{
		Secret:         "test-secret-1",
		ExpirationHour: 1,
	}
	cfg2 := config.JWT{
		Secret:         "test-secret-2",
		ExpirationHour: 1,
	}
	
	jwtManager1 := NewJWTManager(cfg1)
	jwtManager2 := NewJWTManager(cfg2)

	testUserID := uuid.New()
	user := &models.User{
		ID:       testUserID,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	token, err := jwtManager1.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = jwtManager2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for token signed with different secret, got nil")
	}
}

func TestJWTManager_GenerateTokenWithAdminUser(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	testUserID := uuid.New()
	user := &models.User{
		ID:       testUserID,
		Username: "adminuser",
		Email:    "admin@example.com",
		IsAdmin:  true,
	}

	token, err := jwtManager.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}

	if claims.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.IsAdmin != user.IsAdmin {
		t.Errorf("Expected admin status %t, got %t", user.IsAdmin, claims.IsAdmin)
	}

	if !claims.IsAdmin {
		t.Error("Expected admin user to have admin privileges")
	}
}