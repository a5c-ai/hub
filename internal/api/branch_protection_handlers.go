package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// BranchProtectionHandlers contains handlers for branch protection-related endpoints
type BranchProtectionHandlers struct {
	repositoryService services.RepositoryService
	logger           *logrus.Logger
}

// NewBranchProtectionHandlers creates a new branch protection handlers instance
func NewBranchProtectionHandlers(repositoryService services.RepositoryService, logger *logrus.Logger) *BranchProtectionHandlers {
	return &BranchProtectionHandlers{
		repositoryService: repositoryService,
		logger:           logger,
	}
}

// BranchProtection represents branch protection rules
type BranchProtection struct {
	URL                      string                    `json:"url"`
	RequiredStatusChecks     *RequiredStatusChecks     `json:"required_status_checks"`
	RequiredPullRequestReviews *RequiredPullRequestReviews `json:"required_pull_request_reviews"`
	EnforceAdmins            bool                      `json:"enforce_admins"`
	Restrictions             *BranchRestrictions       `json:"restrictions"`
	RequireLinearHistory     bool                      `json:"require_linear_history"`
	AllowForcePushes         bool                      `json:"allow_force_pushes"`
	AllowDeletions           bool                      `json:"allow_deletions"`
	RequireConversationResolution bool                 `json:"require_conversation_resolution"`
}

// RequiredStatusChecks represents required status checks configuration
type RequiredStatusChecks struct {
	Strict   bool     `json:"strict"`
	Contexts []string `json:"contexts"`
}

// RequiredPullRequestReviews represents required pull request reviews configuration
type RequiredPullRequestReviews struct {
	DismissStaleReviews          bool   `json:"dismiss_stale_reviews"`
	RequireCodeOwnerReviews      bool   `json:"require_code_owner_reviews"`
	RequiredApprovingReviewCount int    `json:"required_approving_review_count"`
	RequireLastPushApproval      bool   `json:"require_last_push_approval"`
	DismissalRestrictions        *BranchRestrictions `json:"dismissal_restrictions,omitempty"`
}

// BranchRestrictions represents branch access restrictions
type BranchRestrictions struct {
	Users []gin.H `json:"users"`
	Teams []gin.H `json:"teams"`
	Apps  []gin.H `json:"apps"`
}

// GetBranchProtection handles GET /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection
func (h *BranchProtectionHandlers) GetBranchProtection(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, return mock branch protection data
	// In a full implementation, this would query branch protection rules from the database
	protection := BranchProtection{
		URL: "/api/v1/repositories/" + owner + "/" + repoName + "/branches/" + branch + "/protection",
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"continuous-integration", "security-scan"},
		},
		RequiredPullRequestReviews: &RequiredPullRequestReviews{
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
			RequiredApprovingReviewCount: 2,
			RequireLastPushApproval:      false,
		},
		EnforceAdmins:             true,
		RequireLinearHistory:      false,
		AllowForcePushes:          false,
		AllowDeletions:            false,
		RequireConversationResolution: true,
		Restrictions: &BranchRestrictions{
			Users: []gin.H{},
			Teams: []gin.H{},
			Apps:  []gin.H{},
		},
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Retrieved branch protection rules")

	c.JSON(http.StatusOK, protection)
}

// UpdateBranchProtection handles PUT /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection
func (h *BranchProtectionHandlers) UpdateBranchProtection(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req struct {
		RequiredStatusChecks       *RequiredStatusChecks       `json:"required_status_checks,omitempty"`
		RequiredPullRequestReviews *RequiredPullRequestReviews `json:"required_pull_request_reviews,omitempty"`
		EnforceAdmins              *bool                       `json:"enforce_admins,omitempty"`
		Restrictions               *BranchRestrictions         `json:"restrictions,omitempty"`
		RequireLinearHistory       *bool                       `json:"require_linear_history,omitempty"`
		AllowForcePushes           *bool                       `json:"allow_force_pushes,omitempty"`
		AllowDeletions             *bool                       `json:"allow_deletions,omitempty"`
		RequireConversationResolution *bool                    `json:"require_conversation_resolution,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, return updated mock branch protection data
	// In a full implementation, this would update branch protection rules in the database
	protection := BranchProtection{
		URL: "/api/v1/repositories/" + owner + "/" + repoName + "/branches/" + branch + "/protection",
		RequiredStatusChecks: &RequiredStatusChecks{
			Strict:   true,
			Contexts: []string{"continuous-integration", "security-scan"},
		},
		RequiredPullRequestReviews: &RequiredPullRequestReviews{
			DismissStaleReviews:          true,
			RequireCodeOwnerReviews:      true,
			RequiredApprovingReviewCount: 2,
			RequireLastPushApproval:      false,
		},
		EnforceAdmins:             true,
		RequireLinearHistory:      false,
		AllowForcePushes:          false,
		AllowDeletions:            false,
		RequireConversationResolution: true,
		Restrictions: &BranchRestrictions{
			Users: []gin.H{},
			Teams: []gin.H{},
			Apps:  []gin.H{},
		},
	}

	// Apply updates from request
	if req.RequiredStatusChecks != nil {
		protection.RequiredStatusChecks = req.RequiredStatusChecks
	}
	if req.RequiredPullRequestReviews != nil {
		protection.RequiredPullRequestReviews = req.RequiredPullRequestReviews
	}
	if req.EnforceAdmins != nil {
		protection.EnforceAdmins = *req.EnforceAdmins
	}
	if req.Restrictions != nil {
		protection.Restrictions = req.Restrictions
	}
	if req.RequireLinearHistory != nil {
		protection.RequireLinearHistory = *req.RequireLinearHistory
	}
	if req.AllowForcePushes != nil {
		protection.AllowForcePushes = *req.AllowForcePushes
	}
	if req.AllowDeletions != nil {
		protection.AllowDeletions = *req.AllowDeletions
	}
	if req.RequireConversationResolution != nil {
		protection.RequireConversationResolution = *req.RequireConversationResolution
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Updated branch protection rules")

	c.JSON(http.StatusOK, protection)
}

// DeleteBranchProtection handles DELETE /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection
func (h *BranchProtectionHandlers) DeleteBranchProtection(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, just log the deletion
	// In a full implementation, this would delete branch protection rules from the database
	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Deleted branch protection rules")

	c.JSON(http.StatusNoContent, nil)
}

// GetRequiredStatusChecks handles GET /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_status_checks
func (h *BranchProtectionHandlers) GetRequiredStatusChecks(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, return mock required status checks data
	statusChecks := RequiredStatusChecks{
		Strict:   true,
		Contexts: []string{"continuous-integration", "security-scan"},
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Retrieved required status checks")

	c.JSON(http.StatusOK, statusChecks)
}

// UpdateRequiredStatusChecks handles PATCH /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_status_checks
func (h *BranchProtectionHandlers) UpdateRequiredStatusChecks(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req RequiredStatusChecks
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, return the updated status checks
	// In a full implementation, this would update the status checks in the database
	h.logger.WithFields(logrus.Fields{
		"repo_id":  repo.ID,
		"branch":   branch,
		"strict":   req.Strict,
		"contexts": req.Contexts,
	}).Info("Updated required status checks")

	c.JSON(http.StatusOK, req)
}

// DeleteRequiredStatusChecks handles DELETE /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_status_checks
func (h *BranchProtectionHandlers) DeleteRequiredStatusChecks(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, just log the deletion
	// In a full implementation, this would delete required status checks from the database
	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Deleted required status checks")

	c.JSON(http.StatusNoContent, nil)
}

// GetRequiredPullRequestReviews handles GET /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews
func (h *BranchProtectionHandlers) GetRequiredPullRequestReviews(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, return mock required pull request reviews data
	reviews := RequiredPullRequestReviews{
		DismissStaleReviews:          true,
		RequireCodeOwnerReviews:      true,
		RequiredApprovingReviewCount: 2,
		RequireLastPushApproval:      false,
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Retrieved required pull request reviews")

	c.JSON(http.StatusOK, reviews)
}

// UpdateRequiredPullRequestReviews handles PATCH /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews
func (h *BranchProtectionHandlers) UpdateRequiredPullRequestReviews(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	var req RequiredPullRequestReviews
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, return the updated reviews settings
	// In a full implementation, this would update the settings in the database
	h.logger.WithFields(logrus.Fields{
		"repo_id":                          repo.ID,
		"branch":                           branch,
		"dismiss_stale_reviews":            req.DismissStaleReviews,
		"require_code_owner_reviews":       req.RequireCodeOwnerReviews,
		"required_approving_review_count":  req.RequiredApprovingReviewCount,
		"require_last_push_approval":       req.RequireLastPushApproval,
	}).Info("Updated required pull request reviews")

	c.JSON(http.StatusOK, req)
}

// DeleteRequiredPullRequestReviews handles DELETE /api/v1/repositories/{owner}/{repo}/branches/{branch}/protection/required_pull_request_reviews
func (h *BranchProtectionHandlers) DeleteRequiredPullRequestReviews(c *gin.Context) {
	owner := c.Param("owner")
	repoName := c.Param("repo")
	branch := c.Param("branch")

	if owner == "" || repoName == "" || branch == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Owner, repository name, and branch are required"})
		return
	}

	// Get repository first
	repo, err := h.repositoryService.Get(c.Request.Context(), owner, repoName)
	if err != nil {
		if err.Error() == "repository not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository"})
		}
		return
	}

	// For now, just log the deletion
	// In a full implementation, this would delete required pull request reviews from the database
	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Deleted required pull request reviews")

	c.JSON(http.StatusNoContent, nil)
}