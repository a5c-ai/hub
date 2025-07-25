package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Request/Response Types for Teams
type CreateTeamRequest struct {
	Name         string                `json:"name" binding:"required,min=1,max=255"`
	Description  string                `json:"description,omitempty"`
	Privacy      models.TeamPrivacy    `json:"privacy" binding:"required"`
	ParentTeamID *uuid.UUID            `json:"parent_team_id,omitempty"`
}

type UpdateTeamRequest struct {
	Name         *string               `json:"name,omitempty"`
	Description  *string               `json:"description,omitempty"`
	Privacy      *models.TeamPrivacy   `json:"privacy,omitempty"`
	ParentTeamID *uuid.UUID            `json:"parent_team_id,omitempty"`
}

type TeamFilters struct {
	Privacy *models.TeamPrivacy `json:"privacy,omitempty"`
	Limit   int                 `json:"limit,omitempty"`
	Offset  int                 `json:"offset,omitempty"`
}

// Service Interfaces
type TeamService interface {
	Create(ctx context.Context, orgName string, req CreateTeamRequest) (*models.Team, error)
	Get(ctx context.Context, orgName, teamName string) (*models.Team, error)
	Update(ctx context.Context, orgName, teamName string, req UpdateTeamRequest) (*models.Team, error)
	Delete(ctx context.Context, orgName, teamName string) error
	List(ctx context.Context, orgName string, filters TeamFilters) ([]*models.Team, error)
	GetTeamHierarchy(ctx context.Context, orgName string) ([]*models.Team, error)
}

type TeamMembershipService interface {
	AddMember(ctx context.Context, orgName, teamName, username string, role models.TeamRole) (*models.TeamMember, error)
	RemoveMember(ctx context.Context, orgName, teamName, username string) error
	UpdateMemberRole(ctx context.Context, orgName, teamName, username string, role models.TeamRole) (*models.TeamMember, error)
	GetMembers(ctx context.Context, orgName, teamName string) ([]*models.TeamMember, error)
	GetUserTeams(ctx context.Context, orgName, username string) ([]*models.Team, error)
}

// Team Service Implementation
type teamService struct {
	db *gorm.DB
	as ActivityService
}

func NewTeamService(db *gorm.DB, as ActivityService) TeamService {
	return &teamService{db: db, as: as}
}

func (s *teamService) Create(ctx context.Context, orgName string, req CreateTeamRequest) (*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Validate parent team if specified
	if req.ParentTeamID != nil {
		var parentTeam models.Team
		if err := s.db.Where("id = ? AND organization_id = ?", *req.ParentTeamID, org.ID).First(&parentTeam).Error; err != nil {
			return nil, fmt.Errorf("parent team not found: %w", err)
		}
	}

	team := &models.Team{
		OrganizationID: org.ID,
		Name:           req.Name,
		Description:    req.Description,
		Privacy:        req.Privacy,
		ParentTeamID:   req.ParentTeamID,
	}

	if err := s.db.Create(team).Error; err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Load relationships
	s.db.Preload("Organization").Preload("ParentTeam").First(team, team.ID)

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, uuid.Nil, models.ActivityTeamCreated, "team", &team.ID, map[string]interface{}{
				"name":    team.Name,
				"privacy": team.Privacy,
			})
		}()
	}

	return team, nil
}

func (s *teamService) Get(ctx context.Context, orgName, teamName string) (*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).
		Preload("Organization").
		Preload("ParentTeam").
		Preload("ChildTeams").
		Preload("Members").
		First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	return &team, nil
}

func (s *teamService) Update(ctx context.Context, orgName, teamName string, req UpdateTeamRequest) (*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Privacy != nil {
		updates["privacy"] = *req.Privacy
	}
	if req.ParentTeamID != nil {
		// Validate parent team
		if *req.ParentTeamID != uuid.Nil {
			var parentTeam models.Team
			if err := s.db.Where("id = ? AND organization_id = ?", *req.ParentTeamID, org.ID).First(&parentTeam).Error; err != nil {
				return nil, fmt.Errorf("parent team not found: %w", err)
			}
		}
		updates["parent_team_id"] = req.ParentTeamID
	}

	if err := s.db.Model(&team).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// Load relationships
	s.db.Preload("Organization").Preload("ParentTeam").Preload("ChildTeams").First(&team, team.ID)

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, uuid.Nil, models.ActivityTeamUpdated, "team", &team.ID, map[string]interface{}{
				"updates": updates,
			})
		}()
	}

	return &team, nil
}

func (s *teamService) Delete(ctx context.Context, orgName, teamName string) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	teamID := team.ID

	if err := s.db.Delete(&team).Error; err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, uuid.Nil, models.ActivityTeamDeleted, "team", &teamID, map[string]interface{}{
				"name": teamName,
			})
		}()
	}

	return nil
}

func (s *teamService) List(ctx context.Context, orgName string, filters TeamFilters) ([]*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	query := s.db.Where("organization_id = ?", org.ID).
		Preload("ParentTeam").
		Preload("ChildTeams")

	if filters.Privacy != nil {
		query = query.Where("privacy = ?", *filters.Privacy)
	}
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var teams []*models.Team
	if err := query.Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return teams, nil
}

func (s *teamService) GetTeamHierarchy(ctx context.Context, orgName string) ([]*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get all teams and build hierarchy
	var teams []*models.Team
	if err := s.db.Where("organization_id = ?", org.ID).
		Preload("ChildTeams").
		Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	// Filter to root teams (no parent)
	var rootTeams []*models.Team
	for _, team := range teams {
		if team.ParentTeamID == nil {
			rootTeams = append(rootTeams, team)
		}
	}

	return rootTeams, nil
}

// Team Membership Service Implementation
type teamMembershipService struct {
	db *gorm.DB
	as ActivityService
}

func NewTeamMembershipService(db *gorm.DB, as ActivityService) TeamMembershipService {
	return &teamMembershipService{db: db, as: as}
}

func (s *teamMembershipService) AddMember(ctx context.Context, orgName, teamName, username string, role models.TeamRole) (*models.TeamMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify user is a member of the organization
	var orgMember models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).First(&orgMember).Error; err != nil {
		return nil, fmt.Errorf("user is not a member of the organization: %w", err)
	}

	member := &models.TeamMember{
		TeamID: team.ID,
		UserID: user.ID,
		Role:   role,
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to add team member: %w", err)
	}

	// Load relationships
	s.db.Preload("Team").Preload("User").First(member, member.ID)

	return member, nil
}

func (s *teamMembershipService) RemoveMember(ctx context.Context, orgName, teamName, username string) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.db.Where("team_id = ? AND user_id = ?", team.ID, user.ID).Delete(&models.TeamMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove team member: %w", err)
	}

	return nil
}

func (s *teamMembershipService) UpdateMemberRole(ctx context.Context, orgName, teamName, username string, role models.TeamRole) (*models.TeamMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var member models.TeamMember
	if err := s.db.Where("team_id = ? AND user_id = ?", team.ID, user.ID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("team member not found: %w", err)
	}

	member.Role = role
	if err := s.db.Save(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to update team member role: %w", err)
	}

	// Load relationships
	s.db.Preload("Team").Preload("User").First(&member, member.ID)

	return &member, nil
}

func (s *teamMembershipService) GetMembers(ctx context.Context, orgName, teamName string) ([]*models.TeamMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	var members []*models.TeamMember
	if err := s.db.Where("team_id = ?", team.ID).
		Preload("User").
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}

	return members, nil
}

func (s *teamMembershipService) GetUserTeams(ctx context.Context, orgName, username string) ([]*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var teams []*models.Team
	if err := s.db.Table("teams").
		Joins("JOIN team_members ON teams.id = team_members.team_id").
		Where("teams.organization_id = ? AND team_members.user_id = ?", org.ID, user.ID).
		Preload("ParentTeam").
		Find(&teams).Error; err != nil {
		return nil, fmt.Errorf("failed to get user teams: %w", err)
	}

	return teams, nil
}

// Enhanced Team Service with Templates and Performance Tracking
type TeamTemplateService interface {
	CreateTeamFromTemplate(ctx context.Context, orgName string, templateID uuid.UUID, req CreateTeamFromTemplateRequest) (*models.Team, error)
	GetTeamTemplates(ctx context.Context, orgName string) ([]*models.OrganizationTemplate, error)
	SyncWithExternalSystem(ctx context.Context, orgName, teamName string, externalSystemConfig ExternalSystemConfig) error
	GetTeamPerformanceMetrics(ctx context.Context, orgName, teamName string, period string) (*TeamPerformanceData, error)
	TrackTeamActivity(ctx context.Context, orgName, teamName string, activity TeamActivity) error
}

type CreateTeamFromTemplateRequest struct {
	Name            string            `json:"name" binding:"required,min=1,max=255"`
	Description     string            `json:"description,omitempty"`
	Privacy         models.TeamPrivacy `json:"privacy" binding:"required"`
	ParentTeamID    *uuid.UUID        `json:"parent_team_id,omitempty"`
	TemplateConfig  map[string]interface{} `json:"template_config,omitempty"`
}

type ExternalSystemConfig struct {
	SystemType   string            `json:"system_type"` // "ldap", "active_directory", "okta", etc.
	Config       map[string]interface{} `json:"config"`
	SyncInterval string            `json:"sync_interval"` // "hourly", "daily", "weekly"
	AutoCreate   bool              `json:"auto_create"`
	AutoRemove   bool              `json:"auto_remove"`
}

type TeamPerformanceData struct {
	TeamName        string    `json:"team_name"`
	Period          string    `json:"period"`
	MemberCount     int       `json:"member_count"`
	CommitsCount    int       `json:"commits_count"`
	IssuesResolved  int       `json:"issues_resolved"`
	PullRequests    int       `json:"pull_requests"`
	CodeReviews     int       `json:"code_reviews"`
	AvgResponseTime float64   `json:"avg_response_time_hours"`
	Productivity    float64   `json:"productivity_score"`
	LastUpdated     time.Time `json:"last_updated"`
}

type TeamActivity struct {
	ActivityType string                 `json:"activity_type"`
	ActorID      uuid.UUID              `json:"actor_id"`
	Metadata     map[string]interface{} `json:"metadata"`
	Timestamp    time.Time              `json:"timestamp"`
}

// Team Template Service Implementation
type teamTemplateService struct {
	db *gorm.DB
	as ActivityService
	ts TeamService
}

func NewTeamTemplateService(db *gorm.DB, as ActivityService, ts TeamService) TeamTemplateService {
	return &teamTemplateService{db: db, as: as, ts: ts}
}

func (s *teamTemplateService) CreateTeamFromTemplate(ctx context.Context, orgName string, templateID uuid.UUID, req CreateTeamFromTemplateRequest) (*models.Team, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Get the template
	var template models.OrganizationTemplate
	if err := s.db.Where("id = ? AND organization_id = ? AND template_type = ?", 
		templateID, org.ID, models.TemplateTypeTeam).First(&template).Error; err != nil {
		return nil, fmt.Errorf("team template not found: %w", err)
	}

	// Parse template configuration
	var templateConfig map[string]interface{}
	if err := json.Unmarshal([]byte(template.Configuration), &templateConfig); err != nil {
		return nil, fmt.Errorf("failed to parse template configuration: %w", err)
	}

	// Create team using template
	teamReq := CreateTeamRequest{
		Name:         req.Name,
		Description:  req.Description,
		Privacy:      req.Privacy,
		ParentTeamID: req.ParentTeamID,
	}

	// Apply template defaults if not overridden
	if req.Description == "" {
		if desc, ok := templateConfig["default_description"].(string); ok {
			teamReq.Description = desc
		}
	}

	team, err := s.ts.Create(ctx, orgName, teamReq)
	if err != nil {
		return nil, err
	}

	// Apply template-specific configurations
	if err := s.applyTemplateConfiguration(ctx, team, templateConfig, req.TemplateConfig); err != nil {
		// If configuration fails, we should still return the team but log the error
		if s.as != nil {
			go func() {
				s.as.LogActivity(context.Background(), org.ID, uuid.Nil, models.ActivityAction("team.template_configuration_failed"), "team", &team.ID, map[string]interface{}{
					"template_id": templateID,
					"error":       err.Error(),
				})
			}()
		}
	}

	// Update template usage count
	s.db.Model(&template).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1))

	return team, nil
}

func (s *teamTemplateService) GetTeamTemplates(ctx context.Context, orgName string) ([]*models.OrganizationTemplate, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var templates []*models.OrganizationTemplate
	if err := s.db.Where("organization_id = ? AND template_type = ?", 
		org.ID, models.TemplateTypeTeam).Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to get team templates: %w", err)
	}

	return templates, nil
}

func (s *teamTemplateService) SyncWithExternalSystem(ctx context.Context, orgName, teamName string, config ExternalSystemConfig) error {
	// This would implement integration with external systems like LDAP, Active Directory, etc.
	// For now, it's a placeholder implementation
	return fmt.Errorf("external system sync not yet implemented for system type: %s", config.SystemType)
}

func (s *teamTemplateService) GetTeamPerformanceMetrics(ctx context.Context, orgName, teamName string, period string) (*TeamPerformanceData, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var team models.Team
	if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	// Calculate performance metrics based on period
	var memberCount int64
	s.db.Model(&models.TeamMember{}).Where("team_id = ?", team.ID).Count(&memberCount)

	// These would integrate with actual data sources (git commits, issues, PRs, etc.)
	performance := &TeamPerformanceData{
		TeamName:        teamName,
		Period:          period,
		MemberCount:     int(memberCount),
		CommitsCount:    0,  // Would query git data
		IssuesResolved:  0,  // Would query issue data
		PullRequests:    0,  // Would query PR data
		CodeReviews:     0,  // Would query review data
		AvgResponseTime: 0,  // Would calculate from timestamps
		Productivity:    0,  // Would calculate based on various metrics
		LastUpdated:     time.Now(),
	}

	return performance, nil
}

func (s *teamTemplateService) TrackTeamActivity(ctx context.Context, orgName, teamName string, activity TeamActivity) error {
	// Log team-specific activity for performance tracking
	if s.as != nil {
		var org models.Organization
		if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
			return fmt.Errorf("organization not found: %w", err)
		}

		var team models.Team
		if err := s.db.Where("organization_id = ? AND name = ?", org.ID, teamName).First(&team).Error; err != nil {
			return fmt.Errorf("team not found: %w", err)
		}

		go func() {
			s.as.LogActivity(context.Background(), org.ID, activity.ActorID, 
				models.ActivityAction("team."+activity.ActivityType), "team", &team.ID, activity.Metadata)
		}()
	}

	return nil
}

func (s *teamTemplateService) applyTemplateConfiguration(ctx context.Context, team *models.Team, templateConfig, userConfig map[string]interface{}) error {
	// Apply template-specific configurations such as:
	// - Default repository permissions
	// - Default branch protection rules
	// - Integration settings
	// - Notification preferences
	
	// This is a simplified implementation - would be expanded based on template types
	if permissions, ok := templateConfig["default_permissions"].(map[string]interface{}); ok {
		// Apply default repository permissions for the team
		_ = permissions // Would implement permission application logic
	}

	if integrations, ok := templateConfig["integrations"].(map[string]interface{}); ok {
		// Configure team integrations
		_ = integrations // Would implement integration setup logic
	}

	// Override with user-provided configuration
	for key, value := range userConfig {
		// Apply user overrides
		_ = key
		_ = value
	}

	return nil
}