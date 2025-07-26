package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/a5c-ai/hub/internal/models"
)

func setupWebhookTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.Webhook{},
		&models.WebhookDelivery{},
		&models.DeployKey{},
		&models.WebhookEvent{},
	)
	assert.NoError(t, err)

	return db
}

func TestWebhookDeliveryService_CreateWebhook(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewWebhookDeliveryService(db, logger)

	repositoryID := uuid.New()
	webhook, err := service.CreateWebhook(
		context.Background(),
		repositoryID,
		"test-webhook",
		"https://example.com/webhook",
		"secret123",
		[]string{"push", "pull_request"},
		"application/json",
		false,
		true,
	)

	assert.NoError(t, err)
	assert.NotNil(t, webhook)
	assert.Equal(t, "test-webhook", webhook.Name)
	assert.Equal(t, "https://example.com/webhook", webhook.URL)
	assert.Equal(t, repositoryID, webhook.RepositoryID)
	assert.True(t, webhook.Active)
}

func TestWebhookDeliveryService_ListWebhooks(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewWebhookDeliveryService(db, logger)

	repositoryID := uuid.New()

	// Create two webhooks
	_, err := service.CreateWebhook(context.Background(), repositoryID, "webhook1", "https://example.com/webhook1", "secret1", []string{"push"}, "application/json", false, true)
	assert.NoError(t, err)

	_, err = service.CreateWebhook(context.Background(), repositoryID, "webhook2", "https://example.com/webhook2", "secret2", []string{"pull_request"}, "application/json", false, true)
	assert.NoError(t, err)

	webhooks, err := service.ListWebhooks(context.Background(), repositoryID)
	assert.NoError(t, err)
	assert.Len(t, webhooks, 2)
}

func TestWebhookDeliveryService_VerifySignature(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewWebhookDeliveryService(db, logger)

	secret := "test-secret"
	payload := []byte(`{"test": "payload"}`)

	// Calculate expected signature
	expectedSignature := service.calculateSignature(secret, payload)
	fullSignature := "sha256=" + expectedSignature

	// Test verification
	isValid := service.VerifySignature(secret, fullSignature, payload)
	assert.True(t, isValid)

	// Test with invalid signature
	isValid = service.VerifySignature(secret, "sha256=invalid", payload)
	assert.False(t, isValid)
}

func TestDeployKeyService_CreateDeployKey(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewDeployKeyService(db, logger)

	repositoryID := uuid.New()
	testKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7HtldZp9PV7LQdHEh4hx7YLq5tC7VLyQjhJdPt5yL test@example.com"

	deployKey, err := service.CreateDeployKey(
		context.Background(),
		repositoryID,
		"Test Deploy Key",
		testKey,
		true,
	)

	assert.NoError(t, err)
	assert.NotNil(t, deployKey)
	assert.Equal(t, "Test Deploy Key", deployKey.Title)
	assert.Equal(t, testKey, deployKey.Key)
	assert.Equal(t, repositoryID, deployKey.RepositoryID)
	assert.True(t, deployKey.ReadOnly)
	assert.True(t, deployKey.Verified)
}

func TestDeployKeyService_ValidateSSHKey(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewDeployKeyService(db, logger)

	// Valid SSH key
	validKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7HtldZp9PV7LQdHEh4hx7YLq5tC7VLyQjhJdPt5yL test@example.com"
	err := service.ValidateSSHKey(validKey)
	assert.NoError(t, err)

	// Invalid SSH key
	invalidKey := "invalid-ssh-key"
	err = service.ValidateSSHKey(invalidKey)
	assert.Error(t, err)
}

func TestDeployKeyService_ListDeployKeys(t *testing.T) {
	db := setupWebhookTestDB(t)
	logger := logrus.New()
	service := NewDeployKeyService(db, logger)

	repositoryID := uuid.New()
	testKey1 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7HtldZp9PV7LQdHEh4hx7YLq5tC7VLyQjhJdPt5yL test1@example.com"
	testKey2 := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQD8JtmVZp9PV7LQdHEh4hx7YLq5tC7VLyQjhJdPt5yL test2@example.com"

	// Create two deploy keys
	_, err := service.CreateDeployKey(context.Background(), repositoryID, "Key 1", testKey1, true)
	assert.NoError(t, err)

	_, err = service.CreateDeployKey(context.Background(), repositoryID, "Key 2", testKey2, false)
	assert.NoError(t, err)

	deployKeys, err := service.ListDeployKeys(context.Background(), repositoryID)
	assert.NoError(t, err)
	assert.Len(t, deployKeys, 2)
}
