package auth

import (
	"testing"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/models"
)

func TestJWTManager_GenerateAndValidateToken(t *testing.T) {
	cfg := config.JWT{
		Secret:         "test-secret",
		ExpirationHour: 1,
	}
	jwtManager := NewJWTManager(cfg)

	user := &models.User{
		ID:       1,
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
		t.Errorf("Expected user ID %d, got %d", user.ID, claims.UserID)
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

	user := &models.User{
		ID:       1,
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