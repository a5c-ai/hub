package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WebhookService handles webhook processing for Actions triggers
type WebhookService struct {
	db                  *gorm.DB
	logger              *logrus.Logger
	actionsEventService *ActionsEventService
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB, logger *logrus.Logger, actionsEventService *ActionsEventService) *WebhookService {
	return &WebhookService{
		db:                  db,
		logger:              logger,
		actionsEventService: actionsEventService,
	}
}

// WebhookEvent represents a generic webhook event
type WebhookEvent struct {
	ID          string                 `json:"id"`
	Event       string                 `json:"event"`
	Action      string                 `json:"action,omitempty"`
	Repository  map[string]interface{} `json:"repository"`
	Sender      map[string]interface{} `json:"sender"`
	Payload     map[string]interface{} `json:"payload"`
	Timestamp   time.Time              `json:"timestamp"`
	Signature   string                 `json:"signature,omitempty"`
	DeliveryID  string                 `json:"delivery_id"`
}

// ProcessWebhook processes an incoming webhook and triggers appropriate workflows
func (s *WebhookService) ProcessWebhook(ctx context.Context, headers http.Header, body []byte) error {
	// Extract event type from headers
	eventType := headers.Get("X-Hub-Event")
	if eventType == "" {
		eventType = headers.Get("X-GitHub-Event") // GitHub compatibility
	}
	
	deliveryID := headers.Get("X-Hub-Delivery")
	if deliveryID == "" {
		deliveryID = headers.Get("X-GitHub-Delivery") // GitHub compatibility
	}
	
	signature := headers.Get("X-Hub-Signature-256")
	if signature == "" {
		signature = headers.Get("X-GitHub-Signature-256") // GitHub compatibility
	}

	s.logger.WithFields(logrus.Fields{
		"event_type":  eventType,
		"delivery_id": deliveryID,
		"body_size":   len(body),
	}).Info("Processing webhook")

	// Parse the webhook payload
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Extract repository information
	repository, ok := payload["repository"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("webhook payload missing repository information")
	}

	// Get repository ID from the repository data
	repositoryID, err := s.getRepositoryID(ctx, repository)
	if err != nil {
		return fmt.Errorf("failed to get repository ID: %w", err)
	}

	// Verify webhook signature if provided
	if signature != "" {
		if err := s.verifyWebhookSignature(ctx, repositoryID, signature, body); err != nil {
			return fmt.Errorf("webhook signature verification failed: %w", err)
		}
	}

	// Route to appropriate event handler based on event type
	switch eventType {
	case "push":
		return s.handlePushWebhook(ctx, repositoryID, payload)
	case "pull_request":
		return s.handlePullRequestWebhook(ctx, repositoryID, payload)
	case "issues":
		return s.handleIssuesWebhook(ctx, repositoryID, payload)
	case "issue_comment":
		return s.handleIssueCommentWebhook(ctx, repositoryID, payload)
	case "release":
		return s.handleReleaseWebhook(ctx, repositoryID, payload)
	case "create":
		return s.handleCreateWebhook(ctx, repositoryID, payload)
	case "delete":
		return s.handleDeleteWebhook(ctx, repositoryID, payload)
	case "fork":
		return s.handleForkWebhook(ctx, repositoryID, payload)
	case "watch":
		return s.handleWatchWebhook(ctx, repositoryID, payload)
	case "repository":
		return s.handleRepositoryWebhook(ctx, repositoryID, payload)
	default:
		s.logger.WithField("event_type", eventType).Warn("Unsupported webhook event type")
		return nil // Not an error, just unsupported
	}
}

// handlePushWebhook handles push webhooks
func (s *WebhookService) handlePushWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	// Convert payload to GitPushEvent
	event := GitPushEvent{
		RepositoryID: repositoryID,
	}

	// Extract ref
	if ref, ok := payload["ref"].(string); ok {
		event.Ref = ref
	}

	// Extract before/after SHAs
	if before, ok := payload["before"].(string); ok {
		event.Before = before
	}
	if after, ok := payload["after"].(string); ok {
		event.After = after
	}

	// Extract flags
	if created, ok := payload["created"].(bool); ok {
		event.Created = created
	}
	if deleted, ok := payload["deleted"].(bool); ok {
		event.Deleted = deleted
	}
	if forced, ok := payload["forced"].(bool); ok {
		event.Forced = forced
	}

	// Extract commits
	if commits, ok := payload["commits"].([]interface{}); ok {
		for _, c := range commits {
			if commitMap, ok := c.(map[string]interface{}); ok {
				commit := Commit{}
				if id, ok := commitMap["id"].(string); ok {
					commit.ID = id
				}
				if message, ok := commitMap["message"].(string); ok {
					commit.Message = message
				}
				if added, ok := commitMap["added"].([]interface{}); ok {
					for _, a := range added {
						if path, ok := a.(string); ok {
							commit.Added = append(commit.Added, path)
						}
					}
				}
				if modified, ok := commitMap["modified"].([]interface{}); ok {
					for _, m := range modified {
						if path, ok := m.(string); ok {
							commit.Modified = append(commit.Modified, path)
						}
					}
				}
				if removed, ok := commitMap["removed"].([]interface{}); ok {
					for _, r := range removed {
						if path, ok := r.(string); ok {
							commit.Removed = append(commit.Removed, path)
						}
					}
				}
				event.Commits = append(event.Commits, commit)
			}
		}
	}

	// Extract sender information
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				event.Sender = User{ID: id}
				if login, ok := sender["login"].(string); ok {
					event.Sender.Login = login
				}
			}
		}
	}

	return s.actionsEventService.HandlePushEvent(ctx, event)
}

// handlePullRequestWebhook handles pull request webhooks
func (s *WebhookService) handlePullRequestWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	event := PullRequestEvent{
		RepositoryID: repositoryID,
	}

	// Extract action
	if action, ok := payload["action"].(string); ok {
		event.Action = action
	}

	// Extract pull request number
	if pr, ok := payload["pull_request"].(map[string]interface{}); ok {
		if number, ok := pr["number"].(float64); ok {
			event.Number = int(number)
		}

		// Extract PR details
		event.PullRequest = PullRequest{}
		if id, ok := pr["id"].(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				event.PullRequest.ID = parsed
			}
		}
		if title, ok := pr["title"].(string); ok {
			event.PullRequest.Title = title
		}
		if body, ok := pr["body"].(string); ok {
			event.PullRequest.Body = body
		}

		// Extract head and base
		if head, ok := pr["head"].(map[string]interface{}); ok {
			if ref, ok := head["ref"].(string); ok {
				event.PullRequest.Head.Ref = ref
			}
			if sha, ok := head["sha"].(string); ok {
				event.PullRequest.Head.SHA = sha
			}
		}
		if base, ok := pr["base"].(map[string]interface{}); ok {
			if ref, ok := base["ref"].(string); ok {
				event.PullRequest.Base.Ref = ref
			}
			if sha, ok := base["sha"].(string); ok {
				event.PullRequest.Base.SHA = sha
			}
		}
	}

	// Extract sender
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				event.Sender = User{ID: id}
				if login, ok := sender["login"].(string); ok {
					event.Sender.Login = login
				}
			}
		}
	}

	return s.actionsEventService.HandlePullRequestEvent(ctx, event)
}

// handleIssuesWebhook handles issues webhooks
func (s *WebhookService) handleIssuesWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	event := IssuesEvent{
		RepositoryID: repositoryID,
	}

	// Extract action
	if action, ok := payload["action"].(string); ok {
		event.Action = action
	}

	// Extract issue details
	if issue, ok := payload["issue"].(map[string]interface{}); ok {
		event.Issue = Issue{}
		if id, ok := issue["id"].(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				event.Issue.ID = parsed
			}
		}
		if number, ok := issue["number"].(float64); ok {
			event.Issue.Number = int(number)
		}
		if title, ok := issue["title"].(string); ok {
			event.Issue.Title = title
		}
		if body, ok := issue["body"].(string); ok {
			event.Issue.Body = body
		}
	}

	// Extract sender
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				event.Sender = User{ID: id}
				if login, ok := sender["login"].(string); ok {
					event.Sender.Login = login
				}
			}
		}
	}

	return s.actionsEventService.HandleIssuesEvent(ctx, event)
}

// handleIssueCommentWebhook handles issue comment webhooks
func (s *WebhookService) handleIssueCommentWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	// Extract action
	action := ""
	if a, ok := payload["action"].(string); ok {
		action = a
	}

	// Extract sender ID
	var senderID uuid.UUID
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				senderID = id
			}
		}
	}

	// Create a mock issue comment event and handle it as an issue event
	event := IssuesEvent{
		RepositoryID: repositoryID,
		Action:       action,
		Issue:        Issue{}, // Placeholder
		Sender:       User{ID: senderID},
	}
	return s.actionsEventService.HandleIssuesEvent(ctx, event)
}

// handleReleaseWebhook handles release webhooks
func (s *WebhookService) handleReleaseWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	event := ReleaseEvent{
		RepositoryID: repositoryID,
	}

	// Extract action
	if action, ok := payload["action"].(string); ok {
		event.Action = action
	}

	// Extract release details
	if release, ok := payload["release"].(map[string]interface{}); ok {
		event.Release = Release{}
		if id, ok := release["id"].(string); ok {
			if parsed, err := uuid.Parse(id); err == nil {
				event.Release.ID = parsed
			}
		}
		if tagName, ok := release["tag_name"].(string); ok {
			event.Release.TagName = tagName
		}
		if name, ok := release["name"].(string); ok {
			event.Release.Name = name
		}
	}

	// Extract sender
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				event.Sender = User{ID: id}
			}
		}
	}

	return s.actionsEventService.HandleReleaseEvent(ctx, event)
}

// handleCreateWebhook handles create webhooks (branch/tag creation)
func (s *WebhookService) handleCreateWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	ref := ""
	if r, ok := payload["ref"].(string); ok {
		ref = r
	}

	refType := ""
	if rt, ok := payload["ref_type"].(string); ok {
		refType = rt
	}

	var senderID uuid.UUID
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				senderID = id
			}
		}
	}

	// For create/delete events, we'll handle them as push events
	event := GitPushEvent{
		RepositoryID: repositoryID,
		Ref:          fmt.Sprintf("refs/%ss/%s", refType, ref),
		Created:      true,
		Sender:       User{ID: senderID},
	}
	return s.actionsEventService.HandlePushEvent(ctx, event)
}

// handleDeleteWebhook handles delete webhooks (branch/tag deletion)
func (s *WebhookService) handleDeleteWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	ref := ""
	if r, ok := payload["ref"].(string); ok {
		ref = r
	}

	refType := ""
	if rt, ok := payload["ref_type"].(string); ok {
		refType = rt
	}

	var senderID uuid.UUID
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				senderID = id
			}
		}
	}

	// For delete events, handle as push event with deleted flag
	event := GitPushEvent{
		RepositoryID: repositoryID,
		Ref:          fmt.Sprintf("refs/%ss/%s", refType, ref),
		Deleted:      true,
		Sender:       User{ID: senderID},
	}
	return s.actionsEventService.HandlePushEvent(ctx, event)
}

// handleForkWebhook handles fork webhooks
func (s *WebhookService) handleForkWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	var senderID uuid.UUID
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				senderID = id
			}
		}
	}

	// For fork/watch events, we can log them but they typically don't trigger workflows
	s.logger.WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"sender_id":     senderID,
	}).Info("Fork event received")
	return nil // Fork events typically don't trigger workflows
}

// handleWatchWebhook handles watch webhooks (starring)
func (s *WebhookService) handleWatchWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	action := ""
	if a, ok := payload["action"].(string); ok {
		action = a
	}

	var senderID uuid.UUID
	if sender, ok := payload["sender"].(map[string]interface{}); ok {
		if idStr, ok := sender["id"].(string); ok {
			if id, err := uuid.Parse(idStr); err == nil {
				senderID = id
			}
		}
	}

	// For watch events, we can log them but they typically don't trigger workflows
	s.logger.WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"action":        action,
		"sender_id":     senderID,
	}).Info("Watch event received")
	return nil // Watch events typically don't trigger workflows
}

// handleRepositoryWebhook handles repository webhooks
func (s *WebhookService) handleRepositoryWebhook(ctx context.Context, repositoryID uuid.UUID, payload map[string]interface{}) error {
	// For repository events, we typically don't trigger workflows
	// This is mainly for repository management events
	s.logger.WithField("repository_id", repositoryID).Info("Repository webhook received")
	return nil
}

// getRepositoryID extracts or looks up the repository ID from webhook payload
func (s *WebhookService) getRepositoryID(ctx context.Context, repository map[string]interface{}) (uuid.UUID, error) {
	// Try to get ID directly
	if idStr, ok := repository["id"].(string); ok {
		if id, err := uuid.Parse(idStr); err == nil {
			return id, nil
		}
	}

	// Try to lookup by full name
	if fullName, ok := repository["full_name"].(string); ok {
		parts := strings.Split(fullName, "/")
		if len(parts) == 2 {
			owner, name := parts[0], parts[1]
			
			var repo models.Repository
			err := s.db.WithContext(ctx).
				Where("owner = ? AND name = ?", owner, name).
				First(&repo).Error
			if err != nil {
				return uuid.Nil, fmt.Errorf("repository not found: %s", fullName)
			}
			
			return repo.ID, nil
		}
	}

	return uuid.Nil, fmt.Errorf("unable to determine repository ID from webhook payload")
}

// verifyWebhookSignature verifies the webhook signature
func (s *WebhookService) verifyWebhookSignature(ctx context.Context, repositoryID uuid.UUID, signature string, body []byte) error {
	// Get webhook secret for this repository
	secret, err := s.getWebhookSecret(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("failed to get webhook secret: %w", err)
	}

	// Parse signature
	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature format")
	}
	
	expectedSignature := signature[7:] // Remove "sha256=" prefix
	
	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	computedSignature := hex.EncodeToString(mac.Sum(nil))
	
	// Compare signatures
	if !hmac.Equal([]byte(expectedSignature), []byte(computedSignature)) {
		return fmt.Errorf("signature verification failed")
	}
	
	return nil
}

// getWebhookSecret retrieves the webhook secret for a repository
func (s *WebhookService) getWebhookSecret(ctx context.Context, repositoryID uuid.UUID) (string, error) {
	// In a real implementation, you would store webhook secrets securely
	// For now, return a placeholder
	return "webhook-secret", nil
}

// CreateWebhookURL creates a webhook URL for a repository
func (s *WebhookService) CreateWebhookURL(repositoryID uuid.UUID, baseURL string) string {
	return fmt.Sprintf("%s/api/webhooks/%s", baseURL, repositoryID.String())
}

// TestWebhook sends a test webhook event
func (s *WebhookService) TestWebhook(ctx context.Context, repositoryID uuid.UUID) error {
	testEvent := map[string]interface{}{
		"action": "test",
		"repository": map[string]interface{}{
			"id":        repositoryID.String(),
			"full_name": "test/repo",
		},
		"sender": map[string]interface{}{
			"id":    uuid.New().String(),
			"login": "test-user",
		},
		"timestamp": time.Now(),
	}

	// Convert to JSON
	payload, err := json.Marshal(testEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test webhook: %w", err)
	}

	// Create test headers
	headers := http.Header{}
	headers.Set("X-Hub-Event", "ping")
	headers.Set("X-Hub-Delivery", uuid.New().String())
	headers.Set("Content-Type", "application/json")

	return s.ProcessWebhook(ctx, headers, payload)
}