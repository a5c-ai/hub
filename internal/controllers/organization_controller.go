package controllers

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrganizationController struct {
	orgService        services.OrganizationService
	memberService     services.MembershipService
	invitationService services.InvitationService
	activityService   services.ActivityService
}

func NewOrganizationController(
	orgService services.OrganizationService,
	memberService services.MembershipService,
	invitationService services.InvitationService,
	activityService services.ActivityService,
) *OrganizationController {
	return &OrganizationController{
		orgService:        orgService,
		memberService:     memberService,
		invitationService: invitationService,
		activityService:   activityService,
	}
}

// Organization endpoints
func (ctrl *OrganizationController) CreateOrganization(c *gin.Context) {
	var req services.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ownerID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	org, err := ctrl.orgService.Create(c.Request.Context(), req, ownerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, org)
}

func (ctrl *OrganizationController) GetOrganization(c *gin.Context) {
	orgName := c.Param("org")
	
	org, err := ctrl.orgService.Get(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, org)
}

func (ctrl *OrganizationController) UpdateOrganization(c *gin.Context) {
	orgName := c.Param("org")
	
	var req services.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := ctrl.orgService.Update(c.Request.Context(), orgName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, org)
}

func (ctrl *OrganizationController) DeleteOrganization(c *gin.Context) {
	orgName := c.Param("org")
	
	if err := ctrl.orgService.Delete(c.Request.Context(), orgName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *OrganizationController) ListOrganizations(c *gin.Context) {
	filters := services.OrganizationFilters{}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	orgs, err := ctrl.orgService.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func (ctrl *OrganizationController) GetUserOrganizations(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	orgs, err := ctrl.orgService.GetUserOrganizations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

// Member endpoints
func (ctrl *OrganizationController) GetMembers(c *gin.Context) {
	orgName := c.Param("org")
	
	filters := services.MemberFilters{}
	
	if roleStr := c.Query("role"); roleStr != "" {
		filters.Role = models.OrganizationRole(roleStr)
	}
	
	if publicStr := c.Query("public"); publicStr != "" {
		if public, err := strconv.ParseBool(publicStr); err == nil {
			filters.Public = &public
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	members, err := ctrl.memberService.GetMembers(c.Request.Context(), orgName, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (ctrl *OrganizationController) GetMember(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	member, err := ctrl.memberService.GetMember(c.Request.Context(), orgName, username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	c.JSON(http.StatusOK, member)
}

func (ctrl *OrganizationController) AddMember(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	var req struct {
		Role models.OrganizationRole `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := ctrl.memberService.AddMember(c.Request.Context(), orgName, username, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

func (ctrl *OrganizationController) RemoveMember(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	if err := ctrl.memberService.RemoveMember(c.Request.Context(), orgName, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *OrganizationController) UpdateMemberRole(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	var req struct {
		Role models.OrganizationRole `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := ctrl.memberService.UpdateMemberRole(c.Request.Context(), orgName, username, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

func (ctrl *OrganizationController) SetMemberPublic(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	if err := ctrl.memberService.SetMemberVisibility(c.Request.Context(), orgName, username, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *OrganizationController) SetMemberPrivate(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	if err := ctrl.memberService.SetMemberVisibility(c.Request.Context(), orgName, username, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// Invitation endpoints
func (ctrl *OrganizationController) GetInvitations(c *gin.Context) {
	orgName := c.Param("org")

	invitations, err := ctrl.invitationService.GetPendingInvitations(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": invitations})
}

func (ctrl *OrganizationController) CreateInvitation(c *gin.Context) {
	orgName := c.Param("org")

	var req struct {
		Email string                  `json:"email" binding:"required,email"`
		Role  models.OrganizationRole `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inviterIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	inviterID, err := uuid.Parse(inviterIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	invitation, err := ctrl.invitationService.CreateInvitation(c.Request.Context(), orgName, req.Email, req.Role, inviterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

func (ctrl *OrganizationController) CancelInvitation(c *gin.Context) {
	invitationIDStr := c.Param("invitation_id")
	
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	if err := ctrl.invitationService.CancelInvitation(c.Request.Context(), invitationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *OrganizationController) AcceptInvitation(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := ctrl.invitationService.AcceptInvitation(c.Request.Context(), req.Token, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation accepted successfully"})
}

// Activity endpoints
func (ctrl *OrganizationController) GetActivity(c *gin.Context) {
	orgName := c.Param("org")
	
	limit := 50 // default
	offset := 0 // default
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	activities, err := ctrl.activityService.GetActivity(c.Request.Context(), orgName, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"activities": activities})
}