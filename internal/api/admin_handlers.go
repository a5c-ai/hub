package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminHandlers contains handlers for admin-related endpoints
type AdminHandlers struct {
	authService auth.AuthService
	db          *gorm.DB
	logger      *logrus.Logger
}

// NewAdminHandlers creates a new admin handlers instance
func NewAdminHandlers(authService auth.AuthService, db *gorm.DB, logger *logrus.Logger) *AdminHandlers {
	return &AdminHandlers{
		authService: authService,
		db:          db,
		logger:      logger,
	}
}

// UserQueryParams represents query parameters for user listing and filtering
type UserQueryParams struct {
	Page    int    `form:"page,default=1"`
	PerPage int    `form:"per_page,default=30"`
	Search  string `form:"search"`
	Role    string `form:"role"`
	Status  string `form:"status"`
	SortBy  string `form:"sort_by,default=created_at"`
	SortDir string `form:"sort_dir,default=desc"`
}

// AdminUserResponse represents the user data returned by admin endpoints
type AdminUserResponse struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	FullName         string     `json:"full_name"`
	AvatarURL        string     `json:"avatar_url"`
	Bio              string     `json:"bio"`
	Location         string     `json:"location"`
	Website          string     `json:"website"`
	Company          string     `json:"company"`
	EmailVerified    bool       `json:"email_verified"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	IsActive         bool       `json:"is_active"`
	IsAdmin          bool       `json:"is_admin"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	PhoneNumber      string     `json:"phone_number"`
	Type             string     `json:"type"`
}

// UserUpdateRequest represents the request body for updating a user
type UserUpdateRequest struct {
	FullName    *string `json:"full_name,omitempty"`
	Email       *string `json:"email,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	Location    *string `json:"location,omitempty"`
	Website     *string `json:"website,omitempty"`
	Company     *string `json:"company,omitempty"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
	IsAdmin     *bool   `json:"is_admin,omitempty"`
}

// UserCreateRequest represents the request body for creating a user
type UserCreateRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=12"`
	FullName    string `json:"full_name" binding:"required,min=1,max=255"`
	Bio         string `json:"bio,omitempty"`
	Location    string `json:"location,omitempty"`
	Website     string `json:"website,omitempty"`
	Company     string `json:"company,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	IsAdmin     bool   `json:"is_admin,omitempty"`
}

// toAdminUserResponse converts a user model to admin user response
func toAdminUserResponse(user *models.User) AdminUserResponse {
	return AdminUserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		FullName:         user.FullName,
		AvatarURL:        user.AvatarURL,
		Bio:              user.Bio,
		Location:         user.Location,
		Website:          user.Website,
		Company:          user.Company,
		EmailVerified:    user.EmailVerified,
		TwoFactorEnabled: user.TwoFactorEnabled,
		IsActive:         user.IsActive,
		IsAdmin:          user.IsAdmin,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		LastLoginAt:      user.LastLoginAt,
		PhoneNumber:      user.PhoneNumber,
		Type:             "user",
	}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminHandlers) ListUsers(c *gin.Context) {
	var params UserQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
		return
	}

	// Validate pagination parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 30
	}

	// Validate sort parameters
	validSortFields := []string{"created_at", "updated_at", "username", "email", "full_name", "last_login_at"}
	if !contains(validSortFields, params.SortBy) {
		params.SortBy = "created_at"
	}
	if params.SortDir != "asc" && params.SortDir != "desc" {
		params.SortDir = "desc"
	}

	// Build query
	query := h.db.Model(&models.User{})

	// Apply search filter
	if params.Search != "" {
		searchTerm := "%" + strings.ToLower(params.Search) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(full_name) LIKE ? OR LOWER(company) LIKE ?",
			searchTerm, searchTerm, searchTerm, searchTerm,
		)
	}

	// Apply role filter
	if params.Role == "admin" {
		query = query.Where("is_admin = ?", true)
	} else if params.Role == "user" {
		query = query.Where("is_admin = ?", false)
	}

	// Apply status filter
	if params.Status == "active" {
		query = query.Where("is_active = ?", true)
	} else if params.Status == "inactive" {
		query = query.Where("is_active = ?", false)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	// Apply sorting and pagination
	offset := (params.Page - 1) * params.PerPage
	query = query.Order(params.SortBy + " " + params.SortDir).
		Offset(offset).
		Limit(params.PerPage)

	// Execute query
	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	// Convert to response format
	userResponses := make([]AdminUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = toAdminUserResponse(&user)
	}

	// Calculate pagination info
	totalPages := int((totalCount + int64(params.PerPage) - 1) / int64(params.PerPage))
	hasNext := params.Page < totalPages
	hasPrev := params.Page > 1

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"pagination": gin.H{
			"page":        params.Page,
			"per_page":    params.PerPage,
			"total":       totalCount,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
		"filters": gin.H{
			"search":   params.Search,
			"role":     params.Role,
			"status":   params.Status,
			"sort_by":  params.SortBy,
			"sort_dir": params.SortDir,
		},
	})
}

// GetUser handles GET /api/v1/admin/users/:id
func (h *AdminHandlers) GetUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, toAdminUserResponse(&user))
}

// CreateUser handles POST /api/v1/admin/users
func (h *AdminHandlers) CreateUser(c *gin.Context) {
	var req UserCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	err := h.db.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	}
	if err != gorm.ErrRecordNotFound {
		h.logger.WithError(err).Error("Failed to check existing user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing user"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.WithError(err).Error("Failed to hash password")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create new user
	user := models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Bio:          req.Bio,
		Location:     req.Location,
		Website:      req.Website,
		Company:      req.Company,
		PhoneNumber:  req.PhoneNumber,
		IsActive:     true, // Admin-created users are active by default
		IsAdmin:      req.IsAdmin,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.db.Create(&user).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
	}).Info("Admin created new user")

	c.JSON(http.StatusCreated, toAdminUserResponse(&user))
}

// UpdateUser handles PATCH /api/v1/admin/users/:id
func (h *AdminHandlers) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get existing user
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.Email != nil {
		// Check if email is already taken by another user
		var existingUser models.User
		err := h.db.Where("email = ? AND id != ?", *req.Email, userID).First(&existingUser).Error
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already taken by another user"})
			return
		}
		if err != gorm.ErrRecordNotFound {
			h.logger.WithError(err).Error("Failed to check existing email")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing email"})
			return
		}
		updates["email"] = *req.Email
		updates["email_verified"] = false // Reset verification when email changes
	}
	if req.Bio != nil {
		updates["bio"] = *req.Bio
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Website != nil {
		updates["website"] = *req.Website
	}
	if req.Company != nil {
		updates["company"] = *req.Company
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsAdmin != nil {
		updates["is_admin"] = *req.IsAdmin
	}

	// Update user
	if err := h.db.Model(&user).Where("id = ?", userID).Updates(updates).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Fetch updated user
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"updates":  updates,
	}).Info("Admin updated user")

	c.JSON(http.StatusOK, toAdminUserResponse(&user))
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *AdminHandlers) DeleteUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current admin user to prevent self-deletion
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if currentUserID.(uuid.UUID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Soft delete the user
	if err := h.db.Delete(&user).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"admin_id": currentUserID,
	}).Info("Admin deleted user")

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}

// EnableUser handles POST /api/v1/admin/users/:id/enable
func (h *AdminHandlers) EnableUser(c *gin.Context) {
	h.setUserStatus(c, true)
}

// DisableUser handles POST /api/v1/admin/users/:id/disable
func (h *AdminHandlers) DisableUser(c *gin.Context) {
	h.setUserStatus(c, false)
}

// setUserStatus is a helper function to enable/disable users
func (h *AdminHandlers) setUserStatus(c *gin.Context, isActive bool) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current admin user to prevent self-disabling
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !isActive && currentUserID.(uuid.UUID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot disable your own account"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Update user status
	updates := map[string]interface{}{
		"is_active":  isActive,
		"updated_at": time.Now(),
	}

	if err := h.db.Model(&user).Where("id = ?", userID).Updates(updates).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	action := "disabled"
	if isActive {
		action = "enabled"
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"username":  user.Username,
		"is_active": isActive,
		"admin_id":  currentUserID,
	}).Infof("Admin %s user", action)

	// Fetch updated user
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User " + action + " successfully",
		"user":    toAdminUserResponse(&user),
	})
}

// SetUserRole handles PATCH /api/v1/admin/users/:id/role
func (h *AdminHandlers) SetUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsAdmin bool `json:"is_admin" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get current admin user to prevent self-demotion
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if !req.IsAdmin && currentUserID.(uuid.UUID) == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove admin role from your own account"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// Update user role
	updates := map[string]interface{}{
		"is_admin":   req.IsAdmin,
		"updated_at": time.Now(),
	}

	if err := h.db.Model(&user).Where("id = ?", userID).Updates(updates).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user role")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	role := "user"
	if req.IsAdmin {
		role = "admin"
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": req.IsAdmin,
		"admin_id": currentUserID,
	}).Infof("Admin changed user role to %s", role)

	// Fetch updated user
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get updated user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated to " + role,
		"user":    toAdminUserResponse(&user),
	})
}

// GetUserStats handles GET /api/v1/admin/users/stats
func (h *AdminHandlers) GetUserStats(c *gin.Context) {
	var stats struct {
		TotalUsers      int64 `json:"total_users"`
		ActiveUsers     int64 `json:"active_users"`
		InactiveUsers   int64 `json:"inactive_users"`
		AdminUsers      int64 `json:"admin_users"`
		VerifiedUsers   int64 `json:"verified_users"`
		MFAEnabledUsers int64 `json:"mfa_enabled_users"`
		UsersThisMonth  int64 `json:"users_this_month"`
		UsersLastMonth  int64 `json:"users_last_month"`
		LoginThisWeek   int64 `json:"logins_this_week"`
	}

	// Get total users
	if err := h.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count total users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get active users
	if err := h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count active users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get inactive users
	if err := h.db.Model(&models.User{}).Where("is_active = ?", false).Count(&stats.InactiveUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count inactive users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get admin users
	if err := h.db.Model(&models.User{}).Where("is_admin = ?", true).Count(&stats.AdminUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count admin users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get verified users
	if err := h.db.Model(&models.User{}).Where("email_verified = ?", true).Count(&stats.VerifiedUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count verified users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get MFA enabled users
	if err := h.db.Model(&models.User{}).Where("two_factor_enabled = ?", true).Count(&stats.MFAEnabledUsers).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count MFA enabled users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get users created this month
	thisMonth := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	if err := h.db.Model(&models.User{}).Where("created_at >= ?", thisMonth).Count(&stats.UsersThisMonth).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count users this month")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get users created last month
	lastMonth := thisMonth.AddDate(0, -1, 0)
	lastMonthEnd := thisMonth.Add(-time.Second)
	if err := h.db.Model(&models.User{}).Where("created_at >= ? AND created_at <= ?", lastMonth, lastMonthEnd).Count(&stats.UsersLastMonth).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count users last month")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	// Get users who logged in this week
	thisWeek := time.Now().Truncate(24*time.Hour).AddDate(0, 0, -int(time.Now().Weekday()))
	if err := h.db.Model(&models.User{}).Where("last_login_at >= ?", thisWeek).Count(&stats.LoginThisWeek).Error; err != nil {
		h.logger.WithError(err).Error("Failed to count logins this week")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
