package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Advanced Organization Service Interfaces
type CustomRoleService interface {
	CreateCustomRole(ctx context.Context, orgName string, req CreateCustomRoleRequest) (*models.CustomRole, error)
	GetCustomRole(ctx context.Context, orgName string, roleID uuid.UUID) (*models.CustomRole, error)
	UpdateCustomRole(ctx context.Context, orgName string, roleID uuid.UUID, req UpdateCustomRoleRequest) (*models.CustomRole, error)
	DeleteCustomRole(ctx context.Context, orgName string, roleID uuid.UUID) error
	ListCustomRoles(ctx context.Context, orgName string) ([]*models.CustomRole, error)
	SetDefaultRole(ctx context.Context, orgName string, roleID uuid.UUID) error
}

type OrganizationPolicyService interface {
	CreatePolicy(ctx context.Context, orgName string, req CreatePolicyRequest) (*models.OrganizationPolicy, error)
	GetPolicy(ctx context.Context, orgName string, policyID uuid.UUID) (*models.OrganizationPolicy, error)
	UpdatePolicy(ctx context.Context, orgName string, policyID uuid.UUID, req UpdatePolicyRequest) (*models.OrganizationPolicy, error)
	DeletePolicy(ctx context.Context, orgName string, policyID uuid.UUID) error
	ListPolicies(ctx context.Context, orgName string, policyType *models.PolicyType) ([]*models.OrganizationPolicy, error)
	EnforcePolicy(ctx context.Context, orgName string, policyType models.PolicyType, action string, metadata map[string]interface{}) (bool, error)
}

type OrganizationTemplateService interface {
	CreateTemplate(ctx context.Context, orgName string, req CreateTemplateRequest) (*models.OrganizationTemplate, error)
	GetTemplate(ctx context.Context, orgName string, templateID uuid.UUID) (*models.OrganizationTemplate, error)
	UpdateTemplate(ctx context.Context, orgName string, templateID uuid.UUID, req UpdateTemplateRequest) (*models.OrganizationTemplate, error)
	DeleteTemplate(ctx context.Context, orgName string, templateID uuid.UUID) error
	ListTemplates(ctx context.Context, orgName string, templateType *models.TemplateType) ([]*models.OrganizationTemplate, error)
	UseTemplate(ctx context.Context, orgName string, templateID uuid.UUID) error
}

type OrganizationSettingsService interface {
	GetSettings(ctx context.Context, orgName string) (*models.OrganizationSettings, error)
	UpdateSettings(ctx context.Context, orgName string, req UpdateOrganizationSettingsRequest) (*models.OrganizationSettings, error)
	ValidateIPAccess(ctx context.Context, orgName string, clientIP string) (bool, error)
	CheckComplianceStatus(ctx context.Context, orgName string) (map[string]bool, error)
}

// Request/Response Types
type CreateCustomRoleRequest struct {
	Name        string                 `json:"name" binding:"required,min=1,max=255"`
	Description string                 `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions" binding:"required"`
	Color       string                 `json:"color,omitempty"`
}

type UpdateCustomRoleRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
	Color       *string                `json:"color,omitempty"`
}

type CreatePolicyRequest struct {
	PolicyType    models.PolicyType      `json:"policy_type" binding:"required"`
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	Description   string                 `json:"description,omitempty"`
	Configuration map[string]interface{} `json:"configuration" binding:"required"`
	Enabled       bool                   `json:"enabled"`
	Enforcement   string                 `json:"enforcement,omitempty"`
}

type UpdatePolicyRequest struct {
	Name          *string                `json:"name,omitempty"`
	Description   *string                `json:"description,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	Enabled       *bool                  `json:"enabled,omitempty"`
	Enforcement   *string                `json:"enforcement,omitempty"`
}

type CreateTemplateRequest struct {
	TemplateType  models.TemplateType    `json:"template_type" binding:"required"`
	Name          string                 `json:"name" binding:"required,min=1,max=255"`
	Description   string                 `json:"description,omitempty"`
	Configuration map[string]interface{} `json:"configuration" binding:"required"`
}

type UpdateTemplateRequest struct {
	Name          *string                `json:"name,omitempty"`
	Description   *string                `json:"description,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type UpdateOrganizationSettingsRequest struct {
	// Branding
	PrimaryColor   *string `json:"primary_color,omitempty"`
	SecondaryColor *string `json:"secondary_color,omitempty"`
	LogoURL        *string `json:"logo_url,omitempty"`
	CustomCSS      *string `json:"custom_css,omitempty"`

	// Security Settings
	RequireTwoFactor *bool     `json:"require_two_factor,omitempty"`
	AllowedIPRanges  []string  `json:"allowed_ip_ranges,omitempty"`
	SSOProvider      *string   `json:"sso_provider,omitempty"`
	SSOConfiguration *string   `json:"sso_configuration,omitempty"`
	SessionTimeout   *int      `json:"session_timeout,omitempty"`

	// Repository Settings
	DefaultVisibility         *string `json:"default_visibility,omitempty"`
	AllowPrivateRepos         *bool   `json:"allow_private_repos,omitempty"`
	AllowInternalRepos        *bool   `json:"allow_internal_repos,omitempty"`
	AllowForking              *bool   `json:"allow_forking,omitempty"`
	AllowOutsideCollaborators *bool   `json:"allow_outside_collaborators,omitempty"`

	// Backup and Recovery
	BackupEnabled   *bool   `json:"backup_enabled,omitempty"`
	BackupFrequency *string `json:"backup_frequency,omitempty"`
	RetentionDays   *int    `json:"retention_days,omitempty"`
}

// Service Implementations
type customRoleService struct {
	db *gorm.DB
	as ActivityService
}

func NewCustomRoleService(db *gorm.DB, as ActivityService) CustomRoleService {
	return &customRoleService{db: db, as: as}
}

func (s *customRoleService) CreateCustomRole(ctx context.Context, orgName string, req CreateCustomRoleRequest) (*models.CustomRole, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal permissions: %w", err)
	}

	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	role := &models.CustomRole{
		OrganizationID: org.ID,
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    string(permissionsJSON),
		Color:          color,
	}

	if err := s.db.Create(role).Error; err != nil {
		return nil, fmt.Errorf("failed to create custom role: %w", err)
	}

	return role, nil
}

func (s *customRoleService) GetCustomRole(ctx context.Context, orgName string, roleID uuid.UUID) (*models.CustomRole, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var role models.CustomRole
	if err := s.db.Where("id = ? AND organization_id = ?", roleID, org.ID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("custom role not found: %w", err)
	}

	return &role, nil
}

func (s *customRoleService) UpdateCustomRole(ctx context.Context, orgName string, roleID uuid.UUID, req UpdateCustomRoleRequest) (*models.CustomRole, error) {
	role, err := s.GetCustomRole(ctx, orgName, roleID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Permissions != nil {
		permissionsJSON, err := json.Marshal(req.Permissions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal permissions: %w", err)
		}
		updates["permissions"] = string(permissionsJSON)
	}

	if err := s.db.Model(role).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update custom role: %w", err)
	}

	return role, nil
}

func (s *customRoleService) DeleteCustomRole(ctx context.Context, orgName string, roleID uuid.UUID) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// Check if role is in use
	var count int64
	if err := s.db.Model(&models.OrganizationMember{}).
		Where("custom_role_id = ?", roleID).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete role: %d members are using this role", count)
	}

	if err := s.db.Where("id = ? AND organization_id = ?", roleID, org.ID).Delete(&models.CustomRole{}).Error; err != nil {
		return fmt.Errorf("failed to delete custom role: %w", err)
	}

	return nil
}

func (s *customRoleService) ListCustomRoles(ctx context.Context, orgName string) ([]*models.CustomRole, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var roles []*models.CustomRole
	if err := s.db.Where("organization_id = ?", org.ID).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to list custom roles: %w", err)
	}

	return roles, nil
}

func (s *customRoleService) SetDefaultRole(ctx context.Context, orgName string, roleID uuid.UUID) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Remove default from all roles
		if err := tx.Model(&models.CustomRole{}).
			Where("organization_id = ?", org.ID).
			Update("is_default", false).Error; err != nil {
			return err
		}

		// Set new default
		if err := tx.Model(&models.CustomRole{}).
			Where("id = ? AND organization_id = ?", roleID, org.ID).
			Update("is_default", true).Error; err != nil {
			return err
		}

		return nil
	})
}

// Organization Policy Service Implementation
type organizationPolicyService struct {
	db *gorm.DB
	as ActivityService
}

func NewOrganizationPolicyService(db *gorm.DB, as ActivityService) OrganizationPolicyService {
	return &organizationPolicyService{db: db, as: as}
}

func (s *organizationPolicyService) CreatePolicy(ctx context.Context, orgName string, req CreatePolicyRequest) (*models.OrganizationPolicy, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	configJSON, err := json.Marshal(req.Configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal configuration: %w", err)
	}

	enforcement := req.Enforcement
	if enforcement == "" {
		enforcement = "warn"
	}

	policy := &models.OrganizationPolicy{
		OrganizationID: org.ID,
		PolicyType:     req.PolicyType,
		Name:           req.Name,
		Description:    req.Description,
		Configuration:  string(configJSON),
		Enabled:        req.Enabled,
		Enforcement:    enforcement,
	}

	if err := s.db.Create(policy).Error; err != nil {
		return nil, fmt.Errorf("failed to create policy: %w", err)
	}

	return policy, nil
}

func (s *organizationPolicyService) GetPolicy(ctx context.Context, orgName string, policyID uuid.UUID) (*models.OrganizationPolicy, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var policy models.OrganizationPolicy
	if err := s.db.Where("id = ? AND organization_id = ?", policyID, org.ID).First(&policy).Error; err != nil {
		return nil, fmt.Errorf("policy not found: %w", err)
	}

	return &policy, nil
}

func (s *organizationPolicyService) UpdatePolicy(ctx context.Context, orgName string, policyID uuid.UUID, req UpdatePolicyRequest) (*models.OrganizationPolicy, error) {
	policy, err := s.GetPolicy(ctx, orgName, policyID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Enforcement != nil {
		updates["enforcement"] = *req.Enforcement
	}
	if req.Configuration != nil {
		configJSON, err := json.Marshal(req.Configuration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal configuration: %w", err)
		}
		updates["configuration"] = string(configJSON)
	}

	if err := s.db.Model(policy).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	return policy, nil
}

func (s *organizationPolicyService) DeletePolicy(ctx context.Context, orgName string, policyID uuid.UUID) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	if err := s.db.Where("id = ? AND organization_id = ?", policyID, org.ID).Delete(&models.OrganizationPolicy{}).Error; err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	return nil
}

func (s *organizationPolicyService) ListPolicies(ctx context.Context, orgName string, policyType *models.PolicyType) ([]*models.OrganizationPolicy, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	query := s.db.Where("organization_id = ?", org.ID)
	if policyType != nil {
		query = query.Where("policy_type = ?", *policyType)
	}

	var policies []*models.OrganizationPolicy
	if err := query.Find(&policies).Error; err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	return policies, nil
}

func (s *organizationPolicyService) EnforcePolicy(ctx context.Context, orgName string, policyType models.PolicyType, action string, metadata map[string]interface{}) (bool, error) {
	policies, err := s.ListPolicies(ctx, orgName, &policyType)
	if err != nil {
		return true, err // Allow on error
	}

	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		// Parse configuration and check policy rules
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(policy.Configuration), &config); err != nil {
			continue
		}

		// This is a simplified policy enforcement - in a real implementation,
		// you would have specific policy engines for each policy type
		if violation := s.checkPolicyViolation(policy, action, config, metadata); violation {
			if policy.Enforcement == "block" {
				return false, fmt.Errorf("action blocked by policy: %s", policy.Name)
			}
			// For "warn" enforcement, log but allow
			if s.as != nil {
				go func() {
					s.as.LogActivity(context.Background(), policy.OrganizationID, uuid.Nil, models.ActivityAction("policy.violation"), "policy", &policy.ID, map[string]interface{}{
						"policy_name": policy.Name,
						"action":      action,
						"metadata":    metadata,
					})
				}()
			}
		}
	}

	return true, nil
}

func (s *organizationPolicyService) checkPolicyViolation(policy *models.OrganizationPolicy, action string, config map[string]interface{}, metadata map[string]interface{}) bool {
	// Simplified policy checking - extend this based on policy types
	switch policy.PolicyType {
	case models.PolicyTypeRepositoryCreation:
		if action == "create_repository" {
			// Check naming conventions, visibility restrictions, etc.
			if requiredPrefix, ok := config["required_prefix"].(string); ok {
				if repoName, ok := metadata["name"].(string); ok {
					return !startsWith(repoName, requiredPrefix)
				}
			}
		}
	case models.PolicyTypeMemberInvitation:
		if action == "invite_member" {
			// Check email domain restrictions, etc.
			if allowedDomains, ok := config["allowed_domains"].([]interface{}); ok {
				if email, ok := metadata["email"].(string); ok {
					return !isEmailDomainAllowed(email, allowedDomains)
				}
			}
		}
	}
	return false
}

// Helper functions
func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func isEmailDomainAllowed(email string, allowedDomains []interface{}) bool {
	// Extract domain from email and check against allowed domains
	// Simplified implementation
	return true
}