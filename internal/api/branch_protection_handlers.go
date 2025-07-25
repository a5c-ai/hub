package api

import (
	"encoding/json"
	"net/http"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// BranchProtectionHandlers contains handlers for branch protection-related endpoints
type BranchProtectionHandlers struct {
	repositoryService services.RepositoryService
	branchService     services.BranchService
	logger           *logrus.Logger
}

// NewBranchProtectionHandlers creates a new branch protection handlers instance
func NewBranchProtectionHandlers(repositoryService services.RepositoryService, branchService services.BranchService, logger *logrus.Logger) *BranchProtectionHandlers {
	return &BranchProtectionHandlers{
		repositoryService: repositoryService,
		branchService:     branchService,
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

// convertToServiceStatusChecks converts handler RequiredStatusChecks to service type
func convertToServiceStatusChecks(req *RequiredStatusChecks) *services.RequiredStatusChecks {
	if req == nil {
		return nil
	}
	return &services.RequiredStatusChecks{
		Strict:   req.Strict,
		Contexts: req.Contexts,
	}
}

// convertToServicePRReviews converts handler RequiredPullRequestReviews to service type
func convertToServicePRReviews(req *RequiredPullRequestReviews) *services.RequiredPullRequestReviews {
	if req == nil {
		return nil
	}
	return &services.RequiredPullRequestReviews{
		RequiredApprovingReviewCount: req.RequiredApprovingReviewCount,
		DismissStaleReviews:          req.DismissStaleReviews,
		RequireCodeOwnerReviews:      req.RequireCodeOwnerReviews,
		RestrictPushesToCodeOwners:   false, // Not supported in handler yet
	}
}

// convertToServiceRestrictions converts handler BranchRestrictions to service type
func convertToServiceRestrictions(req *BranchRestrictions) *services.BranchRestrictions {
	if req == nil {
		return nil
	}
	
	var users, teams []string
	for _, user := range req.Users {
		if name, ok := user["login"].(string); ok {
			users = append(users, name)
		}
	}
	for _, team := range req.Teams {
		if name, ok := team["name"].(string); ok {
			teams = append(teams, name)
		}
	}
	
	return &services.BranchRestrictions{
		Users: users,
		Teams: teams,
	}
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

	// Get branch protection rule from database
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Parse JSON fields
	var requiredStatusChecks *RequiredStatusChecks
	var requiredPRReviews *RequiredPullRequestReviews
	var restrictions *BranchRestrictions

	if rule.RequiredStatusChecks != "" {
		if err := json.Unmarshal([]byte(rule.RequiredStatusChecks), &requiredStatusChecks); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required status checks")
		}
	}

	if rule.RequiredPullRequestReviews != "" {
		if err := json.Unmarshal([]byte(rule.RequiredPullRequestReviews), &requiredPRReviews); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required pull request reviews")
		}
	}

	if rule.Restrictions != "" {
		if err := json.Unmarshal([]byte(rule.Restrictions), &restrictions); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal restrictions")
		}
	}

	protection := BranchProtection{
		URL:                        "/api/v1/repositories/" + owner + "/" + repoName + "/branches/" + branch + "/protection",
		RequiredStatusChecks:       requiredStatusChecks,
		RequiredPullRequestReviews: requiredPRReviews,
		EnforceAdmins:              rule.EnforceAdmins,
		RequireLinearHistory:       false, // Not yet implemented in model
		AllowForcePushes:           false, // Not yet implemented in model
		AllowDeletions:             false, // Not yet implemented in model
		RequireConversationResolution: false, // Not yet implemented in model
		Restrictions:               restrictions,
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

	// Check if protection rule already exists for this branch
	existingRule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	
	var rule *models.BranchProtectionRule
	if err != nil && err.Error() == "no protection rule found for branch '"+branch+"'" {
		// Create new protection rule
		createReq := services.CreateBranchProtectionRequest{
			Pattern:                    branch, // Use exact branch name as pattern
			RequiredStatusChecks:       convertToServiceStatusChecks(req.RequiredStatusChecks),
			EnforceAdmins:              req.EnforceAdmins != nil && *req.EnforceAdmins,
			RequiredPullRequestReviews: convertToServicePRReviews(req.RequiredPullRequestReviews),
			Restrictions:               convertToServiceRestrictions(req.Restrictions),
		}
		
		rule, err = h.branchService.CreateProtectionRule(c.Request.Context(), repo.ID, createReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to create branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch protection"})
			return
		}
	} else if err != nil {
		h.logger.WithError(err).Error("Failed to get branch protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		return
	} else {
		// Update existing protection rule
		updateReq := services.UpdateBranchProtectionRequest{
			RequiredStatusChecks:       convertToServiceStatusChecks(req.RequiredStatusChecks),
			EnforceAdmins:              req.EnforceAdmins,
			RequiredPullRequestReviews: convertToServicePRReviews(req.RequiredPullRequestReviews),
			Restrictions:               convertToServiceRestrictions(req.Restrictions),
		}
		
		rule, err = h.branchService.UpdateProtectionRule(c.Request.Context(), existingRule.ID, updateReq)
		if err != nil {
			h.logger.WithError(err).Error("Failed to update branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch protection"})
			return
		}
	}

	// Parse JSON fields for response
	var requiredStatusChecks *RequiredStatusChecks
	var requiredPRReviews *RequiredPullRequestReviews
	var restrictions *BranchRestrictions

	if rule.RequiredStatusChecks != "" {
		if err := json.Unmarshal([]byte(rule.RequiredStatusChecks), &requiredStatusChecks); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required status checks")
		}
	}

	if rule.RequiredPullRequestReviews != "" {
		if err := json.Unmarshal([]byte(rule.RequiredPullRequestReviews), &requiredPRReviews); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required pull request reviews")
		}
	}

	if rule.Restrictions != "" {
		if err := json.Unmarshal([]byte(rule.Restrictions), &restrictions); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal restrictions")
		}
	}

	protection := BranchProtection{
		URL:                        "/api/v1/repositories/" + owner + "/" + repoName + "/branches/" + branch + "/protection",
		RequiredStatusChecks:       requiredStatusChecks,
		RequiredPullRequestReviews: requiredPRReviews,
		EnforceAdmins:              rule.EnforceAdmins,
		RequireLinearHistory:       false, // Not yet implemented in model
		AllowForcePushes:           false, // Not yet implemented in model
		AllowDeletions:             false, // Not yet implemented in model
		RequireConversationResolution: false, // Not yet implemented in model
		Restrictions:               restrictions,
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

	// Get existing protection rule for this branch
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not found"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Delete the protection rule
	if err := h.branchService.DeleteProtectionRule(c.Request.Context(), rule.ID); err != nil {
		h.logger.WithError(err).Error("Failed to delete branch protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete branch protection"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
		"rule_id": rule.ID,
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

	// Get branch protection rule from database
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Parse required status checks from JSON
	var statusChecks RequiredStatusChecks
	if rule.RequiredStatusChecks != "" {
		if err := json.Unmarshal([]byte(rule.RequiredStatusChecks), &statusChecks); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required status checks")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse status checks configuration"})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Required status checks not configured"})
		return
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

	// Get existing protection rule
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Update the status checks
	updateReq := services.UpdateBranchProtectionRequest{
		RequiredStatusChecks: convertToServiceStatusChecks(&req),
	}

	_, err = h.branchService.UpdateProtectionRule(c.Request.Context(), rule.ID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status checks"})
		return
	}

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

	// Get existing protection rule
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Remove status checks by setting to nil
	updateReq := services.UpdateBranchProtectionRequest{
		RequiredStatusChecks: nil,
	}

	_, err = h.branchService.UpdateProtectionRule(c.Request.Context(), rule.ID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove status checks"})
		return
	}

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

	// Get branch protection rule from database
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Parse required pull request reviews from JSON
	var reviews RequiredPullRequestReviews
	if rule.RequiredPullRequestReviews != "" {
		if err := json.Unmarshal([]byte(rule.RequiredPullRequestReviews), &reviews); err != nil {
			h.logger.WithError(err).Error("Failed to unmarshal required pull request reviews")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse pull request reviews configuration"})
			return
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Required pull request reviews not configured"})
		return
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

	// Get existing protection rule
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Update the pull request reviews
	updateReq := services.UpdateBranchProtectionRequest{
		RequiredPullRequestReviews: convertToServicePRReviews(&req),
	}

	_, err = h.branchService.UpdateProtectionRule(c.Request.Context(), rule.ID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pull request reviews"})
		return
	}

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

	// Get existing protection rule
	rule, err := h.branchService.GetProtectionRuleForBranch(c.Request.Context(), repo.ID, branch)
	if err != nil {
		if err.Error() == "no protection rule found for branch '"+branch+"'" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Branch protection not enabled"})
		} else {
			h.logger.WithError(err).Error("Failed to get branch protection rule")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get branch protection"})
		}
		return
	}

	// Remove pull request reviews by setting to nil
	updateReq := services.UpdateBranchProtectionRequest{
		RequiredPullRequestReviews: nil,
	}

	_, err = h.branchService.UpdateProtectionRule(c.Request.Context(), rule.ID, updateReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update protection rule")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove pull request reviews"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"repo_id": repo.ID,
		"branch":  branch,
	}).Info("Deleted required pull request reviews")

	c.JSON(http.StatusNoContent, nil)
}