package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WebhookDeliveryService handles webhook delivery, retry logic, and management
type WebhookDeliveryService struct {
	db     *gorm.DB
	logger *logrus.Logger
	client *http.Client
}

// NewWebhookDeliveryService creates a new webhook delivery service
func NewWebhookDeliveryService(db *gorm.DB, logger *logrus.Logger) *WebhookDeliveryService {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	return &WebhookDeliveryService{
		db:     db,
		logger: logger,
		client: client,
	}
}

// WebhookPayload represents the structure of webhook payload
type WebhookPayload struct {
	Event      string                 `json:"event"`
	Action     string                 `json:"action,omitempty"`
	Repository map[string]interface{} `json:"repository"`
	Sender     map[string]interface{} `json:"sender,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CreateWebhook creates a new webhook configuration
func (s *WebhookDeliveryService) CreateWebhook(ctx context.Context, repositoryID uuid.UUID, name, url, secret string, events []string, contentType string, insecureSSL, active bool) (*models.Webhook, error) {
	webhook := &models.Webhook{
		RepositoryID: repositoryID,
		Name:         name,
		URL:          url,
		Secret:       secret,
		ContentType:  contentType,
		InsecureSSL:  insecureSSL,
		Active:       active,
	}
	
	webhook.SetEventsSlice(events)
	
	if err := s.db.WithContext(ctx).Create(webhook).Error; err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"webhook_id":    webhook.ID,
		"repository_id": repositoryID,
		"url":           url,
		"events":        events,
	}).Info("Created webhook")
	
	return webhook, nil
}

// GetWebhook retrieves a webhook by ID
func (s *WebhookDeliveryService) GetWebhook(ctx context.Context, webhookID uuid.UUID) (*models.Webhook, error) {
	var webhook models.Webhook
	if err := s.db.WithContext(ctx).First(&webhook, "id = ?", webhookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook not found")
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &webhook, nil
}

// ListWebhooks lists all webhooks for a repository
func (s *WebhookDeliveryService) ListWebhooks(ctx context.Context, repositoryID uuid.UUID) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	if err := s.db.WithContext(ctx).Where("repository_id = ?", repositoryID).Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return webhooks, nil
}

// UpdateWebhook updates an existing webhook
func (s *WebhookDeliveryService) UpdateWebhook(ctx context.Context, webhookID uuid.UUID, updates map[string]interface{}) (*models.Webhook, error) {
	var webhook models.Webhook
	if err := s.db.WithContext(ctx).First(&webhook, "id = ?", webhookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("webhook not found")
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	
	if err := s.db.WithContext(ctx).Model(&webhook).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	
	s.logger.WithFields(logrus.Fields{
		"webhook_id": webhookID,
		"updates":    updates,
	}).Info("Updated webhook")
	
	return &webhook, nil
}

// DeleteWebhook deletes a webhook
func (s *WebhookDeliveryService) DeleteWebhook(ctx context.Context, webhookID uuid.UUID) error {
	result := s.db.WithContext(ctx).Delete(&models.Webhook{}, "id = ?", webhookID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete webhook: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("webhook not found")
	}
	
	s.logger.WithField("webhook_id", webhookID).Info("Deleted webhook")
	return nil
}

// TriggerWebhooks triggers all active webhooks for a repository for a specific event
func (s *WebhookDeliveryService) TriggerWebhooks(ctx context.Context, repositoryID uuid.UUID, eventType string, payload map[string]interface{}) error {
	// Get all active webhooks for the repository
	var webhooks []models.Webhook
	if err := s.db.WithContext(ctx).Where("repository_id = ? AND active = ?", repositoryID, true).Find(&webhooks).Error; err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}
	
	if len(webhooks) == 0 {
		s.logger.WithFields(logrus.Fields{
			"repository_id": repositoryID,
			"event_type":    eventType,
		}).Debug("No active webhooks found")
		return nil
	}
	
	// Create webhook event record
	eventData, _ := json.Marshal(payload)
	webhookEvent := &models.WebhookEvent{
		RepositoryID: repositoryID,
		EventType:    eventType,
		EventData:    string(eventData),
		Processed:    false,
	}
	
	if err := s.db.WithContext(ctx).Create(webhookEvent).Error; err != nil {
		s.logger.WithError(err).Error("Failed to create webhook event record")
	}
	
	// Deliver to each webhook
	for _, webhook := range webhooks {
		events := webhook.GetEventsSlice()
		
		// Check if webhook is configured for this event type
		shouldTrigger := false
		for _, event := range events {
			if event == eventType || event == "*" {
				shouldTrigger = true
				break
			}
		}
		
		if !shouldTrigger {
			continue
		}
		
		// Deliver webhook asynchronously
		go func(w models.Webhook) {
			if err := s.DeliverWebhook(context.Background(), w, eventType, payload); err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"webhook_id": w.ID,
					"event_type": eventType,
				}).Error("Failed to deliver webhook")
			}
		}(webhook)
	}
	
	// Mark event as processed
	webhookEvent.Processed = true
	now := time.Now()
	webhookEvent.ProcessedAt = &now
	s.db.WithContext(ctx).Save(webhookEvent)
	
	return nil
}

// DeliverWebhook delivers a single webhook
func (s *WebhookDeliveryService) DeliverWebhook(ctx context.Context, webhook models.Webhook, eventType string, payload map[string]interface{}) error {
	deliveryID := uuid.New().String()
	
	// Create delivery record
	delivery := &models.WebhookDelivery{
		WebhookID:  webhook.ID,
		EventType:  eventType,
		DeliveryID: deliveryID,
		URL:        webhook.URL,
		Attempts:   1,
	}
	
	// Prepare payload
	webhookPayload := WebhookPayload{
		Event:      eventType,
		Repository: payload,
		Timestamp:  time.Now(),
	}
	
	if data, ok := payload["data"]; ok {
		webhookPayload.Data = data.(map[string]interface{})
	}
	if action, ok := payload["action"]; ok {
		webhookPayload.Action = action.(string)
	}
	if sender, ok := payload["sender"]; ok {
		webhookPayload.Sender = sender.(map[string]interface{})
	}
	
	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("Failed to marshal payload: %v", err)
		s.db.WithContext(ctx).Create(delivery)
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	delivery.Payload = string(payloadBytes)
	
	// Attempt delivery
	startTime := time.Now()
	statusCode, responseHeaders, responseBody, err := s.sendWebhookRequest(webhook, deliveryID, payloadBytes)
	duration := time.Since(startTime).Milliseconds()
	
	delivery.Duration = duration
	delivery.StatusCode = statusCode
	delivery.ResponseHeaders = responseHeaders
	delivery.ResponseBody = responseBody
	
	if err != nil {
		delivery.Success = false
		delivery.ErrorMessage = err.Error()
		
		// Schedule retry if it's a retryable error
		if s.isRetryableError(statusCode, err) {
			nextRetry := s.calculateNextRetry(1)
			delivery.NextRetryAt = &nextRetry
		}
	} else if statusCode >= 200 && statusCode < 300 {
		delivery.Success = true
	} else {
		delivery.Success = false
		delivery.ErrorMessage = fmt.Sprintf("HTTP %d: %s", statusCode, responseBody)
		
		if s.isRetryableStatusCode(statusCode) {
			nextRetry := s.calculateNextRetry(1)
			delivery.NextRetryAt = &nextRetry
		}
	}
	
	if err := s.db.WithContext(ctx).Create(delivery).Error; err != nil {
		s.logger.WithError(err).Error("Failed to save webhook delivery")
	}
	
	s.logger.WithFields(logrus.Fields{
		"webhook_id":  webhook.ID,
		"delivery_id": deliveryID,
		"event_type":  eventType,
		"url":         webhook.URL,
		"success":     delivery.Success,
		"status_code": statusCode,
		"duration":    duration,
	}).Info("Webhook delivery completed")
	
	return err
}

// sendWebhookRequest sends the actual HTTP request
func (s *WebhookDeliveryService) sendWebhookRequest(webhook models.Webhook, deliveryID string, payload []byte) (int, string, string, error) {
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", webhook.ContentType)
	req.Header.Set("User-Agent", "Hub-Webhook/1.0")
	req.Header.Set("X-Hub-Event", "push") // This should be dynamic based on event type
	req.Header.Set("X-Hub-Delivery", deliveryID)
	
	// Add HMAC signature if secret is configured
	if webhook.Secret != "" {
		signature := s.calculateSignature(webhook.Secret, payload)
		req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	}
	
	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", "", fmt.Errorf("failed to read response: %w", err)
	}
	
	// Format response headers
	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			responseHeaders[k] = v[0]
		}
	}
	headersJSON, _ := json.Marshal(responseHeaders)
	
	return resp.StatusCode, string(headersJSON), string(responseBody), nil
}

// calculateSignature calculates HMAC-SHA256 signature
func (s *WebhookDeliveryService) calculateSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature verifies webhook signature
func (s *WebhookDeliveryService) VerifySignature(secret, signature string, payload []byte) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	
	expectedSignature := signature[7:] // Remove "sha256=" prefix
	computedSignature := s.calculateSignature(secret, payload)
	
	return hmac.Equal([]byte(expectedSignature), []byte(computedSignature))
}

// isRetryableError determines if an error should trigger a retry
func (s *WebhookDeliveryService) isRetryableError(statusCode int, err error) bool {
	if err != nil {
		// Network errors, timeouts, etc. are retryable
		return true
	}
	
	return s.isRetryableStatusCode(statusCode)
}

// isRetryableStatusCode determines if a status code should trigger a retry
func (s *WebhookDeliveryService) isRetryableStatusCode(statusCode int) bool {
	// Retry on server errors and some client errors
	return statusCode >= 500 || statusCode == 408 || statusCode == 429
}

// calculateNextRetry calculates the next retry time using exponential backoff
func (s *WebhookDeliveryService) calculateNextRetry(attempt int) time.Time {
	// Exponential backoff: 2^attempt minutes, max 24 hours
	backoffMinutes := math.Pow(2, float64(attempt))
	if backoffMinutes > 24*60 {
		backoffMinutes = 24 * 60
	}
	
	return time.Now().Add(time.Duration(backoffMinutes) * time.Minute)
}

// RetryFailedDeliveries retries failed webhook deliveries that are due for retry
func (s *WebhookDeliveryService) RetryFailedDeliveries(ctx context.Context) error {
	var deliveries []models.WebhookDelivery
	now := time.Now()
	
	err := s.db.WithContext(ctx).
		Preload("Webhook").
		Where("success = ? AND next_retry_at IS NOT NULL AND next_retry_at <= ? AND attempts < ?", false, now, 5).
		Find(&deliveries).Error
	
	if err != nil {
		return fmt.Errorf("failed to find failed deliveries: %w", err)
	}
	
	for _, delivery := range deliveries {
		s.logger.WithFields(logrus.Fields{
			"delivery_id": delivery.ID,
			"webhook_id":  delivery.WebhookID,
			"attempt":     delivery.Attempts + 1,
		}).Info("Retrying webhook delivery")
		
		// Parse the original payload
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(delivery.Payload), &payload); err != nil {
			s.logger.WithError(err).Error("Failed to parse payload for retry")
			continue
		}
		
		// Increment attempt counter
		delivery.Attempts++
		
		// Retry the delivery
		startTime := time.Now()
		statusCode, responseHeaders, responseBody, err := s.sendWebhookRequest(delivery.Webhook, delivery.DeliveryID, []byte(delivery.Payload))
		duration := time.Since(startTime).Milliseconds()
		
		delivery.Duration = duration
		delivery.StatusCode = statusCode
		delivery.ResponseHeaders = responseHeaders
		delivery.ResponseBody = responseBody
		delivery.NextRetryAt = nil
		
		if err != nil {
			delivery.Success = false
			delivery.ErrorMessage = err.Error()
			
			// Schedule next retry if attempts < max
			if delivery.Attempts < 5 && s.isRetryableError(statusCode, err) {
				nextRetry := s.calculateNextRetry(delivery.Attempts)
				delivery.NextRetryAt = &nextRetry
			}
		} else if statusCode >= 200 && statusCode < 300 {
			delivery.Success = true
			delivery.ErrorMessage = ""
		} else {
			delivery.Success = false
			delivery.ErrorMessage = fmt.Sprintf("HTTP %d: %s", statusCode, responseBody)
			
			// Schedule next retry if attempts < max
			if delivery.Attempts < 5 && s.isRetryableStatusCode(statusCode) {
				nextRetry := s.calculateNextRetry(delivery.Attempts)
				delivery.NextRetryAt = &nextRetry
			}
		}
		
		if err := s.db.WithContext(ctx).Save(&delivery).Error; err != nil {
			s.logger.WithError(err).Error("Failed to update delivery after retry")
		}
	}
	
	return nil
}

// PingWebhook sends a ping event to test webhook
func (s *WebhookDeliveryService) PingWebhook(ctx context.Context, webhookID uuid.UUID) error {
	webhook, err := s.GetWebhook(ctx, webhookID)
	if err != nil {
		return err
	}
	
	// Create ping payload
	payload := map[string]interface{}{
		"action": "ping",
		"repository": map[string]interface{}{
			"id":        webhook.RepositoryID.String(),
			"full_name": "test/repository",
		},
		"sender": map[string]interface{}{
			"id":    uuid.New().String(),
			"login": "hub-system",
		},
		"timestamp": time.Now(),
	}
	
	return s.DeliverWebhook(ctx, *webhook, "ping", payload)
}

// GetDeliveries gets webhook deliveries for a webhook
func (s *WebhookDeliveryService) GetDeliveries(ctx context.Context, webhookID uuid.UUID, limit, offset int) ([]models.WebhookDelivery, error) {
	var deliveries []models.WebhookDelivery
	
	query := s.db.WithContext(ctx).Where("webhook_id = ?", webhookID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	
	if err := query.Find(&deliveries).Error; err != nil {
		return nil, fmt.Errorf("failed to get deliveries: %w", err)
	}
	
	return deliveries, nil
}