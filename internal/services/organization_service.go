package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Request/Response Types
type CreateOrganizationRequest struct {
	Name         string `json:"name" binding:"required,min=1,max=255"`
	DisplayName  string `json:"display_name" binding:"required,min=1,max=255"`
	Description  string `json:"description,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	Website      string `json:"website,omitempty"`
	Location     string `json:"location,omitempty"`
	Email        string `json:"email,omitempty"`
	BillingEmail string `json:"billing_email,omitempty"`
}

type UpdateOrganizationRequest struct {
	DisplayName  *string `json:"display_name,omitempty"`
	Description  *string `json:"description,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Website      *string `json:"website,omitempty"`
	Location     *string `json:"location,omitempty"`
	Email        *string `json:"email,omitempty"`
	BillingEmail *string `json:"billing_email,omitempty"`
}

type OrganizationFilters struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type MemberFilters struct {
	Role   models.OrganizationRole `json:"role,omitempty"`
	Public *bool                   `json:"public,omitempty"`
	Limit  int                     `json:"limit,omitempty"`
	Offset int                     `json:"offset,omitempty"`
}

// Service Interfaces
type OrganizationService interface {
	Create(ctx context.Context, req CreateOrganizationRequest, ownerID uuid.UUID) (*models.Organization, error)
	Get(ctx context.Context, name string) (*models.Organization, error)
	Update(ctx context.Context, name string, req UpdateOrganizationRequest) (*models.Organization, error)
	Delete(ctx context.Context, name string) error
	List(ctx context.Context, filters OrganizationFilters) ([]*models.Organization, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error)
}

type MembershipService interface {
	AddMember(ctx context.Context, orgName, username string, role models.OrganizationRole) (*models.OrganizationMember, error)
	RemoveMember(ctx context.Context, orgName, username string) error
	UpdateMemberRole(ctx context.Context, orgName, username string, role models.OrganizationRole) (*models.OrganizationMember, error)
	GetMembers(ctx context.Context, orgName string, filters MemberFilters) ([]*models.OrganizationMember, error)
	GetMember(ctx context.Context, orgName, username string) (*models.OrganizationMember, error)
	SetMemberVisibility(ctx context.Context, orgName, username string, public bool) error
}

type InvitationService interface {
	CreateInvitation(ctx context.Context, orgName, email string, role models.OrganizationRole, inviterID uuid.UUID) (*models.OrganizationInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error
	DeclineInvitation(ctx context.Context, token string) error
	GetPendingInvitations(ctx context.Context, orgName string) ([]*models.OrganizationInvitation, error)
	CancelInvitation(ctx context.Context, invitationID uuid.UUID) error
}

type ActivityService interface {
	LogActivity(ctx context.Context, orgID, actorID uuid.UUID, action models.ActivityAction, targetType string, targetID *uuid.UUID, metadata map[string]interface{}) error
	GetActivity(ctx context.Context, orgName string, limit, offset int) ([]*models.OrganizationActivity, error)
}

// Service Implementations
type organizationService struct {
	db *gorm.DB
	as ActivityService
}

func NewOrganizationService(db *gorm.DB, as ActivityService) OrganizationService {
	return &organizationService{db: db, as: as}
}

func (s *organizationService) Create(ctx context.Context, req CreateOrganizationRequest, ownerID uuid.UUID) (*models.Organization, error) {
	org := &models.Organization{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		AvatarURL:    req.AvatarURL,
		Website:      req.Website,
		Location:     req.Location,
		Email:        req.Email,
		BillingEmail: req.BillingEmail,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("failed to create organization: %w", err)
		}

		// Add owner as organization member
		member := &models.OrganizationMember{
			OrganizationID: org.ID,
			UserID:         ownerID,
			Role:           models.OrgRoleOwner,
			PublicMember:   true,
		}

		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("failed to add owner to organization: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log activity (outside transaction to avoid potential issues)
	go func() {
		if s.as != nil {
			s.as.LogActivity(context.Background(), org.ID, ownerID, models.ActivityMemberAdded, "organization", &org.ID, map[string]interface{}{
				"role": models.OrgRoleOwner,
			})
		}
	}()

	return org, nil
}

func (s *organizationService) Get(ctx context.Context, name string) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", name).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}
	return &org, nil
}

func (s *organizationService) Update(ctx context.Context, name string, req UpdateOrganizationRequest) (*models.Organization, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", name).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	updates := make(map[string]interface{})
	if req.DisplayName != nil {
		updates["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.BillingEmail != nil {
		updates["billing_email"] = *req.BillingEmail
	}

	if err := s.db.Model(&org).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return &org, nil
}

func (s *organizationService) Delete(ctx context.Context, name string) error {
	if err := s.db.Where("name = ?", name).Delete(&models.Organization{}).Error; err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (s *organizationService) List(ctx context.Context, filters OrganizationFilters) ([]*models.Organization, error) {
	var orgs []*models.Organization
	query := s.db.Model(&models.Organization{})

	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	if err := query.Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	return orgs, nil
}

func (s *organizationService) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	var orgs []*models.Organization
	if err := s.db.Table("organizations").
		Joins("JOIN organization_members ON organizations.id = organization_members.organization_id").
		Where("organization_members.user_id = ?", userID).
		Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}
	return orgs, nil
}

// Membership Service Implementation
type membershipService struct {
	db *gorm.DB
	as ActivityService
}

func NewMembershipService(db *gorm.DB, as ActivityService) MembershipService {
	return &membershipService{db: db, as: as}
}

func (s *membershipService) AddMember(ctx context.Context, orgName, username string, role models.OrganizationRole) (*models.OrganizationMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	member := &models.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         user.ID,
		Role:           role,
		PublicMember:   false, // Default to private membership
	}

	if err := s.db.Create(member).Error; err != nil {
		return nil, fmt.Errorf("failed to add member: %w", err)
	}

	// Load relationships
	s.db.Preload("Organization").Preload("User").First(member, member.ID)

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, user.ID, models.ActivityMemberAdded, "user", &user.ID, map[string]interface{}{
				"role": role,
			})
		}()
	}

	return member, nil
}

func (s *membershipService) RemoveMember(ctx context.Context, orgName, username string) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.db.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).Delete(&models.OrganizationMember{}).Error; err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, user.ID, models.ActivityMemberRemoved, "user", &user.ID, nil)
		}()
	}

	return nil
}

func (s *membershipService) UpdateMemberRole(ctx context.Context, orgName, username string, role models.OrganizationRole) (*models.OrganizationMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var member models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).First(&member).Error; err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	oldRole := member.Role
	member.Role = role

	if err := s.db.Save(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to update member role: %w", err)
	}

	// Load relationships
	s.db.Preload("Organization").Preload("User").First(&member, member.ID)

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, user.ID, models.ActivityMemberRoleChanged, "user", &user.ID, map[string]interface{}{
				"old_role": oldRole,
				"new_role": role,
			})
		}()
	}

	return &member, nil
}

func (s *membershipService) GetMembers(ctx context.Context, orgName string, filters MemberFilters) ([]*models.OrganizationMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	query := s.db.Where("organization_id = ?", org.ID).Preload("User")

	if filters.Role != "" {
		query = query.Where("role = ?", filters.Role)
	}
	if filters.Public != nil {
		query = query.Where("public_member = ?", *filters.Public)
	}
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var members []*models.OrganizationMember
	if err := query.Find(&members).Error; err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}

	return members, nil
}

func (s *membershipService) GetMember(ctx context.Context, orgName, username string) (*models.OrganizationMember, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var member models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", org.ID, user.ID).
		Preload("Organization").Preload("User").First(&member).Error; err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	return &member, nil
}

func (s *membershipService) SetMemberVisibility(ctx context.Context, orgName, username string, public bool) error {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.db.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND user_id = ?", org.ID, user.ID).
		Update("public_member", public).Error; err != nil {
		return fmt.Errorf("failed to update member visibility: %w", err)
	}

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, user.ID, models.ActivityMemberVisibilityChanged, "user", &user.ID, map[string]interface{}{
				"public": public,
			})
		}()
	}

	return nil
}

// Invitation Service Implementation
type invitationService struct {
	db *gorm.DB
	as ActivityService
}

func NewInvitationService(db *gorm.DB, as ActivityService) InvitationService {
	return &invitationService{db: db, as: as}
}

func (s *invitationService) CreateInvitation(ctx context.Context, orgName, email string, role models.OrganizationRole, inviterID uuid.UUID) (*models.OrganizationInvitation, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	invitation := &models.OrganizationInvitation{
		OrganizationID: org.ID,
		InviterID:      inviterID,
		Email:          email,
		Role:           role,
		Token:          token,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.db.Create(invitation).Error; err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// Load relationships
	s.db.Preload("Organization").Preload("Inviter").First(invitation, invitation.ID)

	// Log activity
	if s.as != nil {
		go func() {
			s.as.LogActivity(context.Background(), org.ID, inviterID, models.ActivityInvitationSent, "invitation", &invitation.ID, map[string]interface{}{
				"email": email,
				"role":  role,
			})
		}()
	}

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	var invitation models.OrganizationInvitation
	if err := s.db.Where("token = ? AND expires_at > ? AND accepted_at IS NULL", token, time.Now()).
		Preload("Organization").First(&invitation).Error; err != nil {
		return fmt.Errorf("invitation not found or expired: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Mark invitation as accepted
		now := time.Now()
		invitation.AcceptedAt = &now
		if err := tx.Save(&invitation).Error; err != nil {
			return fmt.Errorf("failed to update invitation: %w", err)
		}

		// Create organization membership
		member := &models.OrganizationMember{
			OrganizationID: invitation.OrganizationID,
			UserID:         userID,
			Role:           invitation.Role,
			PublicMember:   false,
		}

		if err := tx.Create(member).Error; err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		// Log activity
		if s.as != nil {
			go func() {
				s.as.LogActivity(context.Background(), invitation.OrganizationID, userID, models.ActivityInvitationAccepted, "invitation", &invitation.ID, map[string]interface{}{
					"role": invitation.Role,
				})
			}()
		}

		return nil
	})
}

func (s *invitationService) DeclineInvitation(ctx context.Context, token string) error {
	return s.db.Where("token = ?", token).Delete(&models.OrganizationInvitation{}).Error
}

func (s *invitationService) GetPendingInvitations(ctx context.Context, orgName string) ([]*models.OrganizationInvitation, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	var invitations []*models.OrganizationInvitation
	if err := s.db.Where("organization_id = ? AND expires_at > ? AND accepted_at IS NULL", org.ID, time.Now()).
		Preload("Inviter").Find(&invitations).Error; err != nil {
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	return invitations, nil
}

func (s *invitationService) CancelInvitation(ctx context.Context, invitationID uuid.UUID) error {
	return s.db.Delete(&models.OrganizationInvitation{}, invitationID).Error
}

// Activity Service Implementation
type activityService struct {
	db *gorm.DB
}

func NewActivityService(db *gorm.DB) ActivityService {
	return &activityService{db: db}
}

func (s *activityService) LogActivity(ctx context.Context, orgID, actorID uuid.UUID, action models.ActivityAction, targetType string, targetID *uuid.UUID, metadata map[string]interface{}) error {
	metadataJSON := ""
	if metadata != nil {
		// In a real implementation, you'd properly serialize this to JSON
		// For now, we'll just convert to string representation
		metadataJSON = fmt.Sprintf("%v", metadata)
	}

	activity := &models.OrganizationActivity{
		OrganizationID: orgID,
		ActorID:        actorID,
		Action:         action,
		TargetType:     targetType,
		TargetID:       targetID,
		Metadata:       metadataJSON,
	}

	return s.db.Create(activity).Error
}

func (s *activityService) GetActivity(ctx context.Context, orgName string, limit, offset int) ([]*models.OrganizationActivity, error) {
	var org models.Organization
	if err := s.db.Where("name = ?", orgName).First(&org).Error; err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	query := s.db.Where("organization_id = ?", org.ID).
		Preload("Actor").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var activities []*models.OrganizationActivity
	if err := query.Find(&activities).Error; err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}

// Helper functions
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}