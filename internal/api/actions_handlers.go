package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ActionsHandlers handles Actions-related HTTP requests
type ActionsHandlers struct {
	workflowService     *services.WorkflowService
	runnerService       *services.RunnerService
	repositoryService   services.RepositoryService
	logStreamingService *services.LogStreamingService
	webhookService      *services.WebhookService
	gitService          git.GitService
	logger              *logrus.Logger
}

// NewActionsHandlers creates a new actions handlers instance
func NewActionsHandlers(workflowService *services.WorkflowService, runnerService *services.RunnerService, repositoryService services.RepositoryService, logStreamingService *services.LogStreamingService, webhookService *services.WebhookService, gitService git.GitService, logger *logrus.Logger) *ActionsHandlers {
	return &ActionsHandlers{
		workflowService:     workflowService,
		runnerService:       runnerService,
		repositoryService:   repositoryService,
		logStreamingService: logStreamingService,
		webhookService:      webhookService,
		gitService:          gitService,
		logger:              logger,
	}
}

// ListWorkflows handles GET /api/v1/repos/{owner}/{repo}/actions/workflows
func (h *ActionsHandlers) ListWorkflows(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse query parameters
	enabled := h.parseBoolQuery(c, "enabled")
	limit := h.parseIntQuery(c, "limit", 50)
	offset := h.parseIntQuery(c, "offset", 0)

	req := services.ListWorkflowsRequest{
		RepositoryID: repositoryID,
		Enabled:      enabled,
		Limit:        limit,
		Offset:       offset,
	}

	workflows, total, err := h.workflowService.ListWorkflows(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list workflows")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workflows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflows":   workflows,
		"total_count": total,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetWorkflow handles GET /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}
func (h *ActionsHandlers) GetWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	workflow, err := h.workflowService.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		h.logger.WithError(err).WithField("workflow_id", workflowID).Error("Failed to get workflow")
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// CreateWorkflow handles POST /api/v1/repos/{owner}/{repo}/actions/workflows
func (h *ActionsHandlers) CreateWorkflow(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req services.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.RepositoryID = repositoryID

	workflow, err := h.workflowService.CreateWorkflow(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

// UpdateWorkflow handles PATCH /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}
func (h *ActionsHandlers) UpdateWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var req services.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow, err := h.workflowService.UpdateWorkflow(c.Request.Context(), workflowID, req)
	if err != nil {
		h.logger.WithError(err).WithField("workflow_id", workflowID).Error("Failed to update workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DeleteWorkflow handles DELETE /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}
func (h *ActionsHandlers) DeleteWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	if err := h.workflowService.DeleteWorkflow(c.Request.Context(), workflowID); err != nil {
		h.logger.WithError(err).WithField("workflow_id", workflowID).Error("Failed to delete workflow")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete workflow"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// EnableWorkflow handles PUT /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}/enable
func (h *ActionsHandlers) EnableWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	enabled := true
	req := services.UpdateWorkflowRequest{
		Enabled: &enabled,
	}

	workflow, err := h.workflowService.UpdateWorkflow(c.Request.Context(), workflowID, req)
	if err != nil {
		h.logger.WithError(err).WithField("workflow_id", workflowID).Error("Failed to enable workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DisableWorkflow handles PUT /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}/disable
func (h *ActionsHandlers) DisableWorkflow(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	enabled := false
	req := services.UpdateWorkflowRequest{
		Enabled: &enabled,
	}

	workflow, err := h.workflowService.UpdateWorkflow(c.Request.Context(), workflowID, req)
	if err != nil {
		h.logger.WithError(err).WithField("workflow_id", workflowID).Error("Failed to disable workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

// DispatchWorkflow handles POST /api/v1/repos/{owner}/{repo}/actions/workflows/{workflow_id}/dispatches
func (h *ActionsHandlers) DispatchWorkflow(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflowID, err := uuid.Parse(c.Param("workflow_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow ID"})
		return
	}

	var dispatchReq struct {
		Ref    string                 `json:"ref" binding:"required"`
		Inputs map[string]interface{} `json:"inputs,omitempty"`
	}

	if err := c.ShouldBindJSON(&dispatchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	var actorID *uuid.UUID
	if userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			actorID = &uid
		}
	}

	// Create workflow run
	runReq := services.CreateWorkflowRunRequest{
		WorkflowID: workflowID,
		Event:      "workflow_dispatch",
		HeadSHA:    h.resolveSHAFromRef(c.Request.Context(), repositoryID, dispatchReq.Ref),
		HeadBranch: &dispatchReq.Ref,
		EventPayload: map[string]interface{}{
			"ref":    dispatchReq.Ref,
			"inputs": dispatchReq.Inputs,
		},
		ActorID: actorID,
	}

	run, err := h.workflowService.CreateWorkflowRun(c.Request.Context(), runReq)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"workflow_id":   workflowID,
			"repository_id": repositoryID,
		}).Error("Failed to dispatch workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, run)
}

// ListWorkflowRuns handles GET /api/v1/repos/{owner}/{repo}/actions/runs
func (h *ActionsHandlers) ListWorkflowRuns(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse query parameters
	var workflowID *uuid.UUID
	if wid := c.Query("workflow_id"); wid != "" {
		if parsed, err := uuid.Parse(wid); err == nil {
			workflowID = &parsed
		}
	}

	var status *models.WorkflowRunStatus
	if s := c.Query("status"); s != "" {
		workflowStatus := models.WorkflowRunStatus(s)
		status = &workflowStatus
	}

	var conclusion *models.WorkflowRunConclusion
	if con := c.Query("conclusion"); con != "" {
		workflowConclusion := models.WorkflowRunConclusion(con)
		conclusion = &workflowConclusion
	}

	var event *string
	if e := c.Query("event"); e != "" {
		event = &e
	}

	var actorID *uuid.UUID
	if aid := c.Query("actor_id"); aid != "" {
		if parsed, err := uuid.Parse(aid); err == nil {
			actorID = &parsed
		}
	}

	limit := h.parseIntQuery(c, "limit", 50)
	offset := h.parseIntQuery(c, "offset", 0)

	req := services.ListWorkflowRunsRequest{
		RepositoryID: repositoryID,
		WorkflowID:   workflowID,
		Status:       status,
		Conclusion:   conclusion,
		Event:        event,
		ActorID:      actorID,
		Limit:        limit,
		Offset:       offset,
	}

	runs, total, err := h.workflowService.ListWorkflowRuns(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list workflow runs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workflow runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workflow_runs": runs,
		"total_count":   total,
		"limit":         limit,
		"offset":        offset,
	})
}

// GetWorkflowRun handles GET /api/v1/repos/{owner}/{repo}/actions/runs/{run_id}
func (h *ActionsHandlers) GetWorkflowRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	run, err := h.workflowService.GetWorkflowRun(c.Request.Context(), runID)
	if err != nil {
		h.logger.WithError(err).WithField("run_id", runID).Error("Failed to get workflow run")
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow run not found"})
		return
	}

	c.JSON(http.StatusOK, run)
}

// CancelWorkflowRun handles POST /api/v1/repos/{owner}/{repo}/actions/runs/{run_id}/cancel
func (h *ActionsHandlers) CancelWorkflowRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	if err := h.workflowService.CancelWorkflowRun(c.Request.Context(), runID); err != nil {
		h.logger.WithError(err).WithField("run_id", runID).Error("Failed to cancel workflow run")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow run cancelled"})
}

// RerunWorkflowRun handles POST /api/v1/repos/{owner}/{repo}/actions/runs/{run_id}/rerun
func (h *ActionsHandlers) RerunWorkflowRun(c *gin.Context) {
	runID, err := uuid.Parse(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid run ID"})
		return
	}

	// Get the original run
	originalRun, err := h.workflowService.GetWorkflowRun(c.Request.Context(), runID)
	if err != nil {
		h.logger.WithError(err).WithField("run_id", runID).Error("Failed to get workflow run for rerun")
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow run not found"})
		return
	}

	// Get user ID from context
	userID, _ := c.Get("user_id")
	var actorID *uuid.UUID
	if userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			actorID = &uid
		}
	}

	// Create a new run with the same parameters
	runReq := services.CreateWorkflowRunRequest{
		WorkflowID:   originalRun.WorkflowID,
		Event:        originalRun.Event,
		HeadSHA:      originalRun.HeadSHA,
		HeadBranch:   originalRun.HeadBranch,
		EventPayload: originalRun.EventPayload,
		ActorID:      actorID,
	}

	newRun, err := h.workflowService.CreateWorkflowRun(c.Request.Context(), runReq)
	if err != nil {
		h.logger.WithError(err).WithField("original_run_id", runID).Error("Failed to rerun workflow")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newRun)
}

// DeleteWorkflowRun handles DELETE /api/v1/repos/{owner}/{repo}/actions/runs/{run_id}
func (h *ActionsHandlers) DeleteWorkflowRun(c *gin.Context) {
	// TODO: Implement workflow run deletion
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// Secret Management Endpoints

// ListSecrets handles GET /api/v1/repos/{owner}/{repo}/actions/secrets
func (h *ActionsHandlers) ListSecrets(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secrets, err := h.workflowService.ListSecrets(c.Request.Context(), repositoryID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list secrets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list secrets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"secrets": secrets})
}

// CreateSecret handles POST /api/v1/repos/{owner}/{repo}/actions/secrets
func (h *ActionsHandlers) CreateSecret(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		Name        string  `json:"name" binding:"required"`
		Value       string  `json:"value" binding:"required"`
		Environment *string `json:"environment,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret, err := h.workflowService.CreateSecret(c.Request.Context(), repositoryID, req.Name, req.Value, req.Environment)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create secret")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, secret)
}

// UpdateSecret handles PUT /api/v1/repos/{owner}/{repo}/actions/secrets/{secret_id}
func (h *ActionsHandlers) UpdateSecret(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secretID, err := uuid.Parse(c.Param("secret_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	var req struct {
		Value       string  `json:"value" binding:"required"`
		Environment *string `json:"environment,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secret, err := h.workflowService.UpdateSecret(c.Request.Context(), repositoryID, secretID, req.Value, req.Environment)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update secret")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

// DeleteSecret handles DELETE /api/v1/repos/{owner}/{repo}/actions/secrets/{secret_id}
func (h *ActionsHandlers) DeleteSecret(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secretID, err := uuid.Parse(c.Param("secret_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid secret ID"})
		return
	}

	if err := h.workflowService.DeleteSecret(c.Request.Context(), repositoryID, secretID); err != nil {
		h.logger.WithError(err).Error("Failed to delete secret")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete secret"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Runner Management Endpoints

// ListRunners handles GET /api/v1/repos/{owner}/{repo}/actions/runners
func (h *ActionsHandlers) ListRunners(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := services.ListRunnersRequest{
		RepositoryID: &repositoryID,
	}
	runners, _, err := h.runnerService.ListRunners(c.Request.Context(), req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list runners")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list runners"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"runners": runners})
}

// CreateRunnerRegistrationToken handles POST /api/v1/repos/{owner}/{repo}/actions/runners/registration-token
func (h *ActionsHandlers) CreateRunnerRegistrationToken(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.runnerService.CreateRegistrationToken(c.Request.Context(), &repositoryID, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create registration token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create registration token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

// DeleteRunner handles DELETE /api/v1/repos/{owner}/{repo}/actions/runners/{runner_id}
func (h *ActionsHandlers) DeleteRunner(c *gin.Context) {
	runnerID, err := uuid.Parse(c.Param("runner_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid runner ID"})
		return
	}

	if err := h.runnerService.DeleteRunner(c.Request.Context(), runnerID); err != nil {
		h.logger.WithError(err).Error("Failed to delete runner")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete runner"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Job and Log Endpoints

// GetJobLogs handles GET /api/v1/repos/{owner}/{repo}/actions/jobs/{job_id}/logs
func (h *ActionsHandlers) GetJobLogs(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	logs, err := h.workflowService.GetJobLogs(c.Request.Context(), jobID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get job logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job logs"})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, logs)
}

// StreamJobLogs handles GET /api/v1/repos/{owner}/{repo}/actions/jobs/{job_id}/logs/stream
func (h *ActionsHandlers) StreamJobLogs(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	// Check if client accepts Server-Sent Events
	if c.GetHeader("Accept") != "text/event-stream" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This endpoint requires Accept: text/event-stream"})
		return
	}

	subscriberID := uuid.New().String()
	ctx := c.Request.Context()

	// Set up SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get SSE channel
	sseCh, err := h.logStreamingService.StreamJobLogsSSE(ctx, jobID, subscriberID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to start log stream")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start log stream"})
		return
	}

	// Stream logs
	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case msg, ok := <-sseCh:
			if !ok {
				return false
			}
			c.SSEvent("log", msg)
			return true
		}
	})
}

// GetHistoricalJobLogs handles GET /api/v1/repos/{owner}/{repo}/actions/jobs/{job_id}/logs/history
func (h *ActionsHandlers) GetHistoricalJobLogs(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("job_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	limit := h.parseIntQuery(c, "limit", 100)
	offset := h.parseIntQuery(c, "offset", 0)

	logs, err := h.logStreamingService.GetJobLogs(c.Request.Context(), jobID, limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get historical job logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// GetStepLogs handles GET /api/v1/repos/{owner}/{repo}/actions/steps/{step_id}/logs
func (h *ActionsHandlers) GetStepLogs(c *gin.Context) {
	stepID, err := uuid.Parse(c.Param("step_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid step ID"})
		return
	}

	logs, err := h.logStreamingService.GetStepLogs(c.Request.Context(), stepID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get step logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get step logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// Webhook Endpoints

// HandleWebhook handles POST /api/webhooks/{repository_id}
func (h *ActionsHandlers) HandleWebhook(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		h.logger.WithError(err).Error("Failed to read webhook body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	if err := h.webhookService.ProcessWebhook(c.Request.Context(), c.Request.Header, body); err != nil {
		h.logger.WithError(err).Error("Failed to process webhook")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}

// CreateWebhookURL handles GET /api/v1/repos/{owner}/{repo}/actions/webhooks/url
func (h *ActionsHandlers) CreateWebhookURL(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get base URL from request
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	webhookURL := h.webhookService.CreateWebhookURL(repositoryID, baseURL)
	
	c.JSON(http.StatusOK, gin.H{
		"webhook_url": webhookURL,
		"events": []string{
			"push", "pull_request", "issues", "issue_comment",
			"release", "create", "delete", "fork", "watch",
		},
		"content_type": "application/json",
	})
}

// TestWebhook handles POST /api/v1/repos/{owner}/{repo}/actions/webhooks/test
func (h *ActionsHandlers) TestWebhook(c *gin.Context) {
	repositoryID, err := h.getRepositoryID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.webhookService.TestWebhook(c.Request.Context(), repositoryID); err != nil {
		h.logger.WithError(err).Error("Failed to send test webhook")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send test webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test webhook sent successfully"})
}


// Helper methods

// getRepositoryID extracts repository ID from URL parameters
func (h *ActionsHandlers) getRepositoryID(c *gin.Context) (uuid.UUID, error) {
	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		return uuid.Nil, fmt.Errorf("owner and repo parameters are required")
	}

	// Look up repository by owner/repo name
	repository, err := h.repositoryService.Get(c.Request.Context(), owner, repo)
	if err != nil {
		return uuid.Nil, fmt.Errorf("repository not found: %w", err)
	}

	return repository.ID, nil
}

// parseBoolQuery parses a boolean query parameter
func (h *ActionsHandlers) parseBoolQuery(c *gin.Context, key string) *bool {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return &parsed
		}
	}
	return nil
}

// parseIntQuery parses an integer query parameter with default value
func (h *ActionsHandlers) parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	if value := c.Query(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultValue
}

// resolveSHAFromRef resolves a git reference to its SHA
func (h *ActionsHandlers) resolveSHAFromRef(ctx context.Context, repositoryID uuid.UUID, ref string) string {
	// Get repository path
	repoPath, err := h.repositoryService.GetRepositoryPath(ctx, repositoryID)
	if err != nil {
		h.logger.WithError(err).WithField("repository_id", repositoryID).Error("Failed to get repository path")
		return ref // Return original ref as fallback
	}

	// Resolve SHA using git service
	sha, err := h.gitService.ResolveSHA(ctx, repoPath, ref)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"repository_id": repositoryID,
			"ref":           ref,
		}).Error("Failed to resolve SHA from ref")
		return ref // Return original ref as fallback
	}

	return sha
}
