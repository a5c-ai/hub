package services

import (
	"context"
	"fmt"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission Service Interface
type PermissionService interface {
	GrantRepositoryPermission(ctx context.Context, repoID uuid.UUID, subjectID uuid.UUID, subjectType models.SubjectType, permission models.Permission) error
	RevokeRepositoryPermission(ctx context.Context, repoID uuid.UUID, subjectID uuid.UUID, subjectType models.SubjectType) error
	CheckRepositoryPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID, permission models.Permission) (bool, error)
	GetRepositoryPermissions(ctx context.Context, repoID uuid.UUID) ([]*models.RepositoryPermission, error)
	GetUserRepositoryPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID) (models.Permission, error)
	CalculateUserPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID) (models.Permission, error)
}

// Permission Service Implementation
type permissionService struct {
	db *gorm.DB
	as ActivityService
}

func NewPermissionService(db *gorm.DB, as ActivityService) PermissionService {
	return &permissionService{db: db, as: as}
}

func (s *permissionService) GrantRepositoryPermission(ctx context.Context, repoID uuid.UUID, subjectID uuid.UUID, subjectType models.SubjectType, permission models.Permission) error {
	// Check if permission already exists
	var existing models.RepositoryPermission
	err := s.db.Where("repository_id = ? AND subject_id = ? AND subject_type = ?", repoID, subjectID, subjectType).First(&existing).Error
	
	if err == nil {
		// Update existing permission
		existing.Permission = permission
		if updateErr := s.db.Save(&existing).Error; updateErr != nil {
			return fmt.Errorf("failed to update permission: %w", updateErr)
		}
	} else if err == gorm.ErrRecordNotFound {
		// Create new permission
		newPermission := &models.RepositoryPermission{
			RepositoryID: repoID,
			SubjectID:    subjectID,
			SubjectType:  subjectType,
			Permission:   permission,
		}
		
		if createErr := s.db.Create(newPermission).Error; createErr != nil {
			return fmt.Errorf("failed to create permission: %w", createErr)
		}
	} else {
		return fmt.Errorf("failed to check existing permission: %w", err)
	}

	// Log activity
	if s.as != nil {
		go func() {
			// Get repository to find organization ID
			var repo models.Repository
			if repoErr := s.db.First(&repo, repoID).Error; repoErr == nil && repo.OwnerType == models.OwnerTypeOrganization {
				s.as.LogActivity(context.Background(), repo.OwnerID, subjectID, models.ActivityPermissionGranted, string(subjectType), &subjectID, map[string]interface{}{
					"repository_id": repoID,
					"permission":    permission,
				})
			}
		}()
	}

	return nil
}

func (s *permissionService) RevokeRepositoryPermission(ctx context.Context, repoID uuid.UUID, subjectID uuid.UUID, subjectType models.SubjectType) error {
	result := s.db.Where("repository_id = ? AND subject_id = ? AND subject_type = ?", repoID, subjectID, subjectType).Delete(&models.RepositoryPermission{})
	
	if result.Error != nil {
		return fmt.Errorf("failed to revoke permission: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}

	// Log activity
	if s.as != nil {
		go func() {
			// Get repository to find organization ID
			var repo models.Repository
			if repoErr := s.db.First(&repo, repoID).Error; repoErr == nil && repo.OwnerType == models.OwnerTypeOrganization {
				s.as.LogActivity(context.Background(), repo.OwnerID, subjectID, models.ActivityPermissionRevoked, string(subjectType), &subjectID, map[string]interface{}{
					"repository_id": repoID,
				})
			}
		}()
	}

	return nil
}

func (s *permissionService) CheckRepositoryPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID, permission models.Permission) (bool, error) {
	userPermission, err := s.CalculateUserPermission(ctx, userID, repoID)
	if err != nil {
		return false, err
	}

	return isPermissionSufficient(userPermission, permission), nil
}

func (s *permissionService) GetRepositoryPermissions(ctx context.Context, repoID uuid.UUID) ([]*models.RepositoryPermission, error) {
	var permissions []*models.RepositoryPermission
	if err := s.db.Where("repository_id = ?", repoID).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get repository permissions: %w", err)
	}
	return permissions, nil
}

func (s *permissionService) GetUserRepositoryPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID) (models.Permission, error) {
	var permission models.RepositoryPermission
	if err := s.db.Where("repository_id = ? AND subject_id = ? AND subject_type = ?", repoID, userID, models.SubjectTypeUser).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // No direct permission
		}
		return "", fmt.Errorf("failed to get user permission: %w", err)
	}
	return permission.Permission, nil
}

func (s *permissionService) CalculateUserPermission(ctx context.Context, userID uuid.UUID, repoID uuid.UUID) (models.Permission, error) {
	// Get repository information
	var repo models.Repository
	if err := s.db.First(&repo, repoID).Error; err != nil {
		return "", fmt.Errorf("repository not found: %w", err)
	}

	// 1. Check if user owns the repository (personal repo)
	if repo.OwnerType == models.OwnerTypeUser && repo.OwnerID == userID {
		return models.PermissionAdmin, nil
	}

	// 2. Check direct user permission
	directPerm, err := s.GetUserRepositoryPermission(ctx, userID, repoID)
	if err != nil {
		return "", err
	}
	if directPerm != "" {
		return directPerm, nil
	}

	// 3. For organization repositories, check organization and team permissions
	if repo.OwnerType == models.OwnerTypeOrganization {
		orgPermission, err := s.calculateOrganizationPermission(ctx, userID, repo.OwnerID, repoID)
		if err != nil {
			return "", err
		}
		if orgPermission != "" {
			return orgPermission, nil
		}
	}

	// 4. Check public repository access
	if repo.Visibility == models.VisibilityPublic {
		return models.PermissionRead, nil
	}

	// 5. Check internal repository access for organization members
	if repo.Visibility == models.VisibilityInternal && repo.OwnerType == models.OwnerTypeOrganization {
		var orgMember models.OrganizationMember
		if err := s.db.Where("organization_id = ? AND user_id = ?", repo.OwnerID, userID).First(&orgMember).Error; err == nil {
			return models.PermissionRead, nil
		}
	}

	// No permission found
	return "", nil
}

func (s *permissionService) calculateOrganizationPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, repoID uuid.UUID) (models.Permission, error) {
	// Check if user is organization owner/admin
	var orgMember models.OrganizationMember
	if err := s.db.Where("organization_id = ? AND user_id = ?", orgID, userID).First(&orgMember).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil // Not an organization member
		}
		return "", fmt.Errorf("failed to check organization membership: %w", err)
	}

	// Organization owners and admins have admin access to all repos
	if orgMember.Role == models.OrgRoleOwner || orgMember.Role == models.OrgRoleAdmin {
		return models.PermissionAdmin, nil
	}

	// Check team permissions
	teamPerm, err := s.getHighestTeamPermission(ctx, userID, orgID, repoID)
	if err != nil {
		return "", err
	}

	return teamPerm, nil
}

func (s *permissionService) getHighestTeamPermission(ctx context.Context, userID uuid.UUID, orgID uuid.UUID, repoID uuid.UUID) (models.Permission, error) {
	// Get all teams the user belongs to in this organization
	var teamMembers []models.TeamMember
	if err := s.db.Table("team_members").
		Joins("JOIN teams ON team_members.team_id = teams.id").
		Where("teams.organization_id = ? AND team_members.user_id = ?", orgID, userID).
		Find(&teamMembers).Error; err != nil {
		return "", fmt.Errorf("failed to get user teams: %w", err)
	}

	var highestPermission models.Permission
	
	for _, teamMember := range teamMembers {
		// Get repository permissions for this team
		var repoPermission models.RepositoryPermission
		if err := s.db.Where("repository_id = ? AND subject_id = ? AND subject_type = ?", 
			repoID, teamMember.TeamID, models.SubjectTypeTeam).First(&repoPermission).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return "", fmt.Errorf("failed to get team permission: %w", err)
			}
			continue // No permission found for this team
		}

		// Check if this is the highest permission so far
		if isHigherPermission(repoPermission.Permission, highestPermission) {
			highestPermission = repoPermission.Permission
		}
	}

	return highestPermission, nil
}

// Helper functions for permission comparison
func isHigherPermission(perm1, perm2 models.Permission) bool {
	if perm2 == "" {
		return perm1 != ""
	}
	
	permissionLevels := map[models.Permission]int{
		models.PermissionRead:     1,
		models.PermissionTriage:   2,
		models.PermissionWrite:    3,
		models.PermissionMaintain: 4,
		models.PermissionAdmin:    5,
	}

	return permissionLevels[perm1] > permissionLevels[perm2]
}

func isPermissionSufficient(userPerm, requiredPerm models.Permission) bool {
	if userPerm == "" {
		return false
	}
	
	permissionLevels := map[models.Permission]int{
		models.PermissionRead:     1,
		models.PermissionTriage:   2,
		models.PermissionWrite:    3,
		models.PermissionMaintain: 4,
		models.PermissionAdmin:    5,
	}

	return permissionLevels[userPerm] >= permissionLevels[requiredPerm]
}