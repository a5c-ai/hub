package controllers

import (
	"net/http"
	"strconv"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamController struct {
	teamService           services.TeamService
	teamMembershipService services.TeamMembershipService
	permissionService     services.PermissionService
}

func NewTeamController(
	teamService services.TeamService,
	teamMembershipService services.TeamMembershipService,
	permissionService services.PermissionService,
) *TeamController {
	return &TeamController{
		teamService:           teamService,
		teamMembershipService: teamMembershipService,
		permissionService:     permissionService,
	}
}

// Team management endpoints
func (ctrl *TeamController) ListTeams(c *gin.Context) {
	orgName := c.Param("org")
	
	filters := services.TeamFilters{}
	
	if privacyStr := c.Query("privacy"); privacyStr != "" {
		privacy := models.TeamPrivacy(privacyStr)
		filters.Privacy = &privacy
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

	teams, err := ctrl.teamService.List(c.Request.Context(), orgName, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (ctrl *TeamController) CreateTeam(c *gin.Context) {
	orgName := c.Param("org")
	
	var req services.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := ctrl.teamService.Create(c.Request.Context(), orgName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, team)
}

func (ctrl *TeamController) GetTeam(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")

	team, err := ctrl.teamService.Get(c.Request.Context(), orgName, teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (ctrl *TeamController) UpdateTeam(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	
	var req services.UpdateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := ctrl.teamService.Update(c.Request.Context(), orgName, teamName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, team)
}

func (ctrl *TeamController) DeleteTeam(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")

	if err := ctrl.teamService.Delete(c.Request.Context(), orgName, teamName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *TeamController) GetTeamHierarchy(c *gin.Context) {
	orgName := c.Param("org")

	teams, err := ctrl.teamService.GetTeamHierarchy(c.Request.Context(), orgName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// Team membership endpoints
func (ctrl *TeamController) GetTeamMembers(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")

	members, err := ctrl.teamMembershipService.GetMembers(c.Request.Context(), orgName, teamName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (ctrl *TeamController) AddTeamMember(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	username := c.Param("username")

	var req struct {
		Role models.TeamRole `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := ctrl.teamMembershipService.AddMember(c.Request.Context(), orgName, teamName, username, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

func (ctrl *TeamController) RemoveTeamMember(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	username := c.Param("username")

	if err := ctrl.teamMembershipService.RemoveMember(c.Request.Context(), orgName, teamName, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (ctrl *TeamController) UpdateTeamMemberRole(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	username := c.Param("username")

	var req struct {
		Role models.TeamRole `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member, err := ctrl.teamMembershipService.UpdateMemberRole(c.Request.Context(), orgName, teamName, username, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, member)
}

func (ctrl *TeamController) GetUserTeams(c *gin.Context) {
	orgName := c.Param("org")
	username := c.Param("username")

	teams, err := ctrl.teamMembershipService.GetUserTeams(c.Request.Context(), orgName, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

// Team repository permissions endpoints
func (ctrl *TeamController) GetTeamRepositories(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")

	// Get team first to get team ID
	team, err := ctrl.teamService.Get(c.Request.Context(), orgName, teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	permissions, err := ctrl.permissionService.GetRepositoryPermissions(c.Request.Context(), team.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"repositories": permissions})
}

func (ctrl *TeamController) AddTeamRepository(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	repoName := c.Param("repo")

	var req struct {
		Permission models.Permission `json:"permission" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get team first to get team ID
	team, err := ctrl.teamService.Get(c.Request.Context(), orgName, teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// For now, we'll assume the repository ID is provided as repoName
	// In a real implementation, you'd look up the repository by name
	repoID, err := uuid.Parse(repoName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	if err := ctrl.permissionService.GrantRepositoryPermission(c.Request.Context(), repoID, team.ID, models.SubjectTypeTeam, req.Permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Repository permission granted"})
}

func (ctrl *TeamController) RemoveTeamRepository(c *gin.Context) {
	orgName := c.Param("org")
	teamName := c.Param("team")
	repoName := c.Param("repo")

	// Get team first to get team ID
	team, err := ctrl.teamService.Get(c.Request.Context(), orgName, teamName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// For now, we'll assume the repository ID is provided as repoName
	// In a real implementation, you'd look up the repository by name
	repoID, err := uuid.Parse(repoName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	if err := ctrl.permissionService.RevokeRepositoryPermission(c.Request.Context(), repoID, team.ID, models.SubjectTypeTeam); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}