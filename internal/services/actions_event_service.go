package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/actions"
	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ActionsEventService handles workflow triggering based on repository events
type ActionsEventService struct {
	db                *gorm.DB
	workflowService   *WorkflowService
	repositoryService RepositoryService
	gitService        git.GitService
	logger            *logrus.Logger
}

// NewActionsEventService creates a new actions event service
func NewActionsEventService(db *gorm.DB, workflowService *WorkflowService, repositoryService RepositoryService, gitService git.GitService, logger *logrus.Logger) *ActionsEventService {
	return &ActionsEventService{
		db:                db,
		workflowService:   workflowService,
		repositoryService: repositoryService,
		gitService:        gitService,
		logger:            logger,
	}
}

// GitPushEvent represents a git push event
type GitPushEvent struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	Ref          string    `json:"ref"`          // refs/heads/main, refs/tags/v1.0.0
	Before       string    `json:"before"`       // Previous commit SHA
	After        string    `json:"after"`        // New commit SHA
	Created      bool      `json:"created"`      // true if this is a new branch/tag
	Deleted      bool      `json:"deleted"`      // true if branch/tag was deleted
	Forced       bool      `json:"forced"`       // true if this was a force push
	Compare      string    `json:"compare"`      // Compare URL
	Commits      []Commit  `json:"commits"`      // List of commits in the push
	HeadCommit   *Commit   `json:"head_commit"`  // The most recent commit
	Pusher       User      `json:"pusher"`       // User who pushed
	Sender       User      `json:"sender"`       // User who sent the event
}

// PullRequestEvent represents a pull request event
type PullRequestEvent struct {
	RepositoryID  uuid.UUID   `json:"repository_id"`
	Action        string      `json:"action"`        // opened, closed, synchronize, reopened, etc.
	Number        int         `json:"number"`        // PR number
	PullRequest   PullRequest `json:"pull_request"`  // PR details
	Changes       interface{} `json:"changes"`       // Changes made (for edited events)
	RequestedReviewer *User   `json:"requested_reviewer"` // For review_requested events
	Sender        User        `json:"sender"`        // User who triggered the event
}

// IssuesEvent represents an issues event
type IssuesEvent struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	Action       string    `json:"action"`       // opened, closed, reopened, etc.
	Issue        Issue     `json:"issue"`        // Issue details
	Changes      interface{} `json:"changes"`    // Changes made
	Assignee     *User     `json:"assignee"`     // For assigned/unassigned events
	Label        *Label    `json:"label"`        // For labeled/unlabeled events
	Sender       User      `json:"sender"`       // User who triggered the event
}

// ReleaseEvent represents a release event
type ReleaseEvent struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	Action       string    `json:"action"`   // created, edited, published, unpublished, deleted
	Release      Release   `json:"release"`  // Release details
	Changes      interface{} `json:"changes"` // Changes made
	Sender       User      `json:"sender"`   // User who triggered the event
}

// Event types for simplified structs (these would normally come from models package)
type Commit struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    Author    `json:"author"`
	Committer Author    `json:"committer"`
	Added     []string  `json:"added"`
	Removed   []string  `json:"removed"`
	Modified  []string  `json:"modified"`
}

type Author struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type User struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	AvatarURL string   `json:"avatar_url"`
}

type PullRequest struct {
	ID     uuid.UUID `json:"id"`
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Body   string    `json:"body"`
	State  string    `json:"state"`
	Head   PRBranch  `json:"head"`
	Base   PRBranch  `json:"base"`
	User   User      `json:"user"`
}

type PRBranch struct {
	Ref string `json:"ref"`
	SHA string `json:"sha"`
}

type Issue struct {
	ID     uuid.UUID `json:"id"`
	Number int       `json:"number"`
	Title  string    `json:"title"`
	Body   string    `json:"body"`
	State  string    `json:"state"`
	User   User      `json:"user"`
}

type Release struct {
	ID      uuid.UUID `json:"id"`
	TagName string    `json:"tag_name"`
	Name    string    `json:"name"`
	Body    string    `json:"body"`
	Draft   bool      `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

type Label struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// HandlePushEvent handles git push events and triggers appropriate workflows
func (s *ActionsEventService) HandlePushEvent(ctx context.Context, event GitPushEvent) error {
	s.logger.WithFields(logrus.Fields{
		"repository_id": event.RepositoryID,
		"ref":           event.Ref,
		"after":         event.After,
		"commits":       len(event.Commits),
	}).Info("Handling push event")

	// Don't trigger on deleted branches/tags unless they have specific workflows for that
	if event.Deleted {
		s.logger.WithField("ref", event.Ref).Info("Skipping deleted ref")
		return nil
	}

	// Extract branch/tag name from ref
	var branch, tag string
	if strings.HasPrefix(event.Ref, "refs/heads/") {
		branch = strings.TrimPrefix(event.Ref, "refs/heads/")
	} else if strings.HasPrefix(event.Ref, "refs/tags/") {
		tag = strings.TrimPrefix(event.Ref, "refs/tags/")
	} else {
		s.logger.WithField("ref", event.Ref).Warn("Unknown ref format")
		return nil
	}

	// Create trigger context
	triggerCtx := actions.TriggerContext{
		Branch: branch,
		Tag:    tag,
		Paths:  s.extractModifiedPaths(event.Commits),
	}

	// Find workflows that should be triggered
	workflows, err := s.workflowService.CheckWorkflowTrigger(ctx, event.RepositoryID, "push", triggerCtx)
	if err != nil {
		return fmt.Errorf("failed to check workflow triggers: %w", err)
	}

	// Trigger each matching workflow
	for _, workflow := range workflows {
		if err := s.triggerWorkflow(ctx, workflow, "push", event.After, &branch, event); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"workflow_id":   workflow.ID,
				"repository_id": event.RepositoryID,
			}).Error("Failed to trigger workflow")
			// Continue with other workflows even if one fails
		}
	}

	return nil
}

// HandlePullRequestEvent handles pull request events and triggers appropriate workflows
func (s *ActionsEventService) HandlePullRequestEvent(ctx context.Context, event PullRequestEvent) error {
	s.logger.WithFields(logrus.Fields{
		"repository_id": event.RepositoryID,
		"action":        event.Action,
		"pr_number":     event.Number,
	}).Info("Handling pull request event")

	// Create trigger context
	triggerCtx := actions.TriggerContext{
		PRAction:       event.Action,
		PRTargetBranch: event.PullRequest.Base.Ref,
		// TODO: Add paths from PR diff
	}

	// Find workflows that should be triggered
	workflows, err := s.workflowService.CheckWorkflowTrigger(ctx, event.RepositoryID, "pull_request", triggerCtx)
	if err != nil {
		return fmt.Errorf("failed to check workflow triggers: %w", err)
	}

	// Trigger each matching workflow
	for _, workflow := range workflows {
		if err := s.triggerWorkflow(ctx, workflow, "pull_request", event.PullRequest.Head.SHA, &event.PullRequest.Head.Ref, event); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"workflow_id":   workflow.ID,
				"repository_id": event.RepositoryID,
			}).Error("Failed to trigger workflow")
		}
	}

	return nil
}

// HandleIssuesEvent handles issues events and triggers appropriate workflows
func (s *ActionsEventService) HandleIssuesEvent(ctx context.Context, event IssuesEvent) error {
	s.logger.WithFields(logrus.Fields{
		"repository_id": event.RepositoryID,
		"action":        event.Action,
		"issue_number":  event.Issue.Number,
	}).Info("Handling issues event")

	// Create trigger context
	triggerCtx := actions.TriggerContext{}

	// Find workflows that should be triggered
	workflows, err := s.workflowService.CheckWorkflowTrigger(ctx, event.RepositoryID, "issues", triggerCtx)
	if err != nil {
		return fmt.Errorf("failed to check workflow triggers: %w", err)
	}

	// For issues events, we use the default branch's HEAD as the SHA
	defaultBranch, defaultSHA := s.getRepositoryDefaultBranchAndSHA(ctx, event.RepositoryID)

	// Trigger each matching workflow
	for _, workflow := range workflows {
		if err := s.triggerWorkflow(ctx, workflow, "issues", defaultSHA, &defaultBranch, event); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"workflow_id":   workflow.ID,
				"repository_id": event.RepositoryID,
			}).Error("Failed to trigger workflow")
		}
	}

	return nil
}

// HandleReleaseEvent handles release events and triggers appropriate workflows
func (s *ActionsEventService) HandleReleaseEvent(ctx context.Context, event ReleaseEvent) error {
	s.logger.WithFields(logrus.Fields{
		"repository_id": event.RepositoryID,
		"action":        event.Action,
		"tag_name":      event.Release.TagName,
	}).Info("Handling release event")

	// Only trigger on certain release actions
	if event.Action != "published" && event.Action != "created" {
		s.logger.WithField("action", event.Action).Info("Skipping release event action")
		return nil
	}

	// Create trigger context
	triggerCtx := actions.TriggerContext{
		Tag: event.Release.TagName,
	}

	// Find workflows that should be triggered
	workflows, err := s.workflowService.CheckWorkflowTrigger(ctx, event.RepositoryID, "release", triggerCtx)
	if err != nil {
		return fmt.Errorf("failed to check workflow triggers: %w", err)
	}

	// For releases, use the tag as the ref
	tagRef := event.Release.TagName
	tagSHA := s.resolveSHAFromRef(ctx, event.RepositoryID, tagRef)

	// Trigger each matching workflow
	for _, workflow := range workflows {
		if err := s.triggerWorkflow(ctx, workflow, "release", tagSHA, &tagRef, event); err != nil {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"workflow_id":   workflow.ID,
				"repository_id": event.RepositoryID,
			}).Error("Failed to trigger workflow")
		}
	}

	return nil
}

// HandleScheduledEvent handles scheduled (cron) workflow triggers
func (s *ActionsEventService) HandleScheduledEvent(ctx context.Context, repositoryID uuid.UUID, workflowID uuid.UUID) error {
	s.logger.WithFields(logrus.Fields{
		"repository_id": repositoryID,
		"workflow_id":   workflowID,
	}).Info("Handling scheduled event")

	// Get the workflow
	workflow, err := s.workflowService.GetWorkflow(ctx, workflowID)
	if err != nil {
		return fmt.Errorf("workflow not found: %w", err)
	}

	// Use the default branch for scheduled runs
	defaultBranch, defaultSHA := s.getRepositoryDefaultBranchAndSHA(ctx, repositoryID)

	return s.triggerWorkflow(ctx, *workflow, "schedule", defaultSHA, &defaultBranch, map[string]interface{}{
		"schedule": "cron",
	})
}

// triggerWorkflow creates a new workflow run for the given workflow and event
func (s *ActionsEventService) triggerWorkflow(ctx context.Context, workflow models.Workflow, event, headSHA string, headBranch *string, eventPayload interface{}) error {
	// Extract actor ID from event payload if available
	var actorID *uuid.UUID
	if eventMap, ok := eventPayload.(map[string]interface{}); ok {
		if sender, exists := eventMap["sender"]; exists {
			if senderMap, ok := sender.(map[string]interface{}); ok {
				if id, exists := senderMap["id"]; exists {
					if idStr, ok := id.(string); ok {
						if parsed, err := uuid.Parse(idStr); err == nil {
							actorID = &parsed
						}
					}
				}
			}
		}
	}

	// Create workflow run request
	runReq := CreateWorkflowRunRequest{
		WorkflowID:   workflow.ID,
		Event:        event,
		HeadSHA:      headSHA,
		HeadBranch:   headBranch,
		EventPayload: eventPayload,
		ActorID:      actorID,
	}

	run, err := s.workflowService.CreateWorkflowRun(ctx, runReq)
	if err != nil {
		return fmt.Errorf("failed to create workflow run: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"workflow_run_id": run.ID,
		"workflow_id":     workflow.ID,
		"repository_id":   workflow.RepositoryID,
		"event":           event,
		"head_sha":        headSHA,
		"run_number":      run.Number,
	}).Info("Created workflow run")

	return nil
}

// extractModifiedPaths extracts all modified file paths from commits
func (s *ActionsEventService) extractModifiedPaths(commits []Commit) []string {
	pathMap := make(map[string]bool)
	
	for _, commit := range commits {
		for _, path := range commit.Added {
			pathMap[path] = true
		}
		for _, path := range commit.Modified {
			pathMap[path] = true
		}
		for _, path := range commit.Removed {
			pathMap[path] = true
		}
	}

	paths := make([]string, 0, len(pathMap))
	for path := range pathMap {
		paths = append(paths, path)
	}

	return paths
}

// resolveSHAFromRef resolves a git reference to its SHA
func (s *ActionsEventService) resolveSHAFromRef(ctx context.Context, repositoryID uuid.UUID, ref string) string {
	// Get repository path
	repoPath, err := s.repositoryService.GetRepositoryPath(ctx, repositoryID)
	if err != nil {
		s.logger.WithError(err).WithField("repository_id", repositoryID).Error("Failed to get repository path")
		return "HEAD" // Return HEAD as fallback
	}

	// Resolve SHA using git service
	sha, err := s.gitService.ResolveSHA(ctx, repoPath, ref)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"repository_id": repositoryID,
			"ref":           ref,
		}).Error("Failed to resolve SHA from ref")
		return "HEAD" // Return HEAD as fallback
	}

	return sha
}

// getRepositoryDefaultBranchAndSHA gets the default branch name and its current SHA
func (s *ActionsEventService) getRepositoryDefaultBranchAndSHA(ctx context.Context, repositoryID uuid.UUID) (string, string) {
	// Get repository path
	repoPath, err := s.repositoryService.GetRepositoryPath(ctx, repositoryID)
	if err != nil {
		s.logger.WithError(err).WithField("repository_id", repositoryID).Error("Failed to get repository path")
		return "main", "HEAD" // Return defaults as fallback
	}

	// Get repository info to find default branch
	repoInfo, err := s.gitService.GetRepositoryInfo(ctx, repoPath)
	if err != nil {
		s.logger.WithError(err).WithField("repository_id", repositoryID).Error("Failed to get repository info")
		return "main", "HEAD" // Return defaults as fallback
	}

	// Get SHA for the default branch
	sha, err := s.gitService.ResolveSHA(ctx, repoPath, repoInfo.DefaultBranch)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"repository_id":    repositoryID,
			"default_branch":   repoInfo.DefaultBranch,
		}).Error("Failed to resolve SHA for default branch")
		return repoInfo.DefaultBranch, "HEAD" // Return branch with HEAD as fallback
	}

	return repoInfo.DefaultBranch, sha
}

