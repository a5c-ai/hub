package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// UserHandlers contains handlers for user-related endpoints
// UserHandlers contains handlers for user-related endpoints
type UserHandlers struct {
	authService         auth.AuthService
	db                  *gorm.DB
	config              *config.Config
	logger              *logrus.Logger
	notificationService services.NotificationService
}

// NewUserHandlers creates a new user handlers instance
// NewUserHandlers creates a new user handlers instance
func NewUserHandlers(
	authService auth.AuthService,
	db *gorm.DB,
	cfg *config.Config,
	logger *logrus.Logger,
	notificationService services.NotificationService,
) *UserHandlers {
	return &UserHandlers{
		authService:         authService,
		db:                  db,
		config:              cfg,
		logger:              logger,
		notificationService: notificationService,
	}
}

// GetUserProfile handles GET /api/v1/users/{username}
func (h *UserHandlers) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	user, err := h.authService.GetUserByUsername(username)
	if err != nil {
		h.logger.WithError(err).WithField("username", username).Error("Failed to get user profile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return public user profile information
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"full_name":  user.FullName,
		"avatar_url": user.AvatarURL,
		"bio":        user.Bio,
		"company":    user.Company,
		"location":   user.Location,
		"website":    user.Website,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
		"type":       "user",
	})
}

// GetCurrentUserProfile handles GET /api/v1/user
func (h *UserHandlers) GetCurrentUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return full user profile information (including private fields)
	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"full_name":      user.FullName,
		"avatar_url":     user.AvatarURL,
		"bio":            user.Bio,
		"company":        user.Company,
		"location":       user.Location,
		"website":        user.Website,
		"email_verified": user.EmailVerified,
		"mfa_enabled":    user.TwoFactorEnabled,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"type":           "user",
	})
}

// UpdateUserProfile handles PATCH /api/v1/user
func (h *UserHandlers) UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		FullName  *string `json:"full_name,omitempty"`
		Bio       *string `json:"bio,omitempty"`
		Company   *string `json:"company,omitempty"`
		Location  *string `json:"location,omitempty"`
		Website   *string `json:"website,omitempty"`
		AvatarURL *string `json:"avatar_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Get current user
	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields if provided
	if req.FullName != nil {
		user.FullName = *req.FullName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Company != nil {
		user.Company = *req.Company
	}
	if req.Location != nil {
		user.Location = *req.Location
	}
	if req.Website != nil {
		user.Website = *req.Website
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}

	// Update user in database
	if err := h.authService.UpdateUser(user); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	// Return updated user profile
	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"username":       user.Username,
		"email":          user.Email,
		"full_name":      user.FullName,
		"avatar_url":     user.AvatarURL,
		"bio":            user.Bio,
		"company":        user.Company,
		"location":       user.Location,
		"website":        user.Website,
		"email_verified": user.EmailVerified,
		"mfa_enabled":    user.TwoFactorEnabled,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"type":           "user",
	})
}

// GetUserRepositories handles GET /api/v1/users/{username}/repositories
func (h *UserHandlers) GetUserRepositories(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Get user first
	_, err := h.authService.GetUserByUsername(username)
	if err != nil {
		h.logger.WithError(err).WithField("username", username).Error("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// For now, return empty repositories list
	// In a full implementation, this would query repositories owned by the user
	c.JSON(http.StatusOK, []gin.H{})
}

// GetUserOrganizations handles GET /api/v1/users/{username}/organizations
func (h *UserHandlers) GetUserOrganizations(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	// Get user first
	_, err := h.authService.GetUserByUsername(username)
	if err != nil {
		h.logger.WithError(err).WithField("username", username).Error("Failed to get user")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// For now, return empty organizations list
	// In a full implementation, this would query organizations the user belongs to
	c.JSON(http.StatusOK, []gin.H{})
}

// GetUserActivity handles GET /api/v1/user/activity
func (h *UserHandlers) GetUserActivity(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, return empty activity feed
	// In a full implementation, this would query user's activity from the activity service
	c.JSON(http.StatusOK, gin.H{
		"activities": []gin.H{},
		"pagination": gin.H{
			"page":     1,
			"per_page": 30,
			"total":    0,
			"has_more": false,
		},
	})
}

// GetNotifications handles GET /api/v1/notifications
func (h *UserHandlers) GetNotifications(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	participating := c.Query("participating") == "true"
	all := c.Query("all") == "true"

	// For now, return empty notifications list
	// In a full implementation, this would query notifications from the database
	c.JSON(http.StatusOK, gin.H{
		"notifications": []gin.H{},
		"pagination": gin.H{
			"page":     1,
			"per_page": 30,
			"total":    0,
			"has_more": false,
		},
		"filters": gin.H{
			"participating": participating,
			"all":           all,
		},
	})
}

// MarkNotificationsAsRead handles PATCH /api/v1/notifications
func (h *UserHandlers) MarkNotificationsAsRead(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		LastReadAt string `json:"last_read_at,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, just return success
	// In a full implementation, this would update notification read status
	c.JSON(http.StatusOK, gin.H{
		"message": "Notifications marked as read",
	})
}

// SubscribeNotifications upgrades connection to WebSocket and streams real-time notifications
func (h *UserHandlers) SubscribeNotifications(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.WithError(err).Error("WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	ch, cancel := h.notificationService.Subscribe(userID)
	defer cancel()

	for notif := range ch {
		if err := conn.WriteJSON(notif); err != nil {
			h.logger.WithError(err).Error("Failed to write notification to WebSocket")
			break
		}
	}
}

// GetEmailVerificationStatus handles GET /api/v1/user/email/verification-status
func (h *UserHandlers) GetEmailVerificationStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Initialize email verification service
	emailService := auth.NewSMTPEmailService(h.config)
	verificationService := auth.NewEmailVerificationService(h.db, emailService)

	// Get verification status and any active token
	verified, token, err := verificationService.GetVerificationStatus(userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get email verification status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get verification status"})
		return
	}

	response := gin.H{
		"verified": verified,
	}

	if !verified && token != nil {
		response["pending_token"] = gin.H{
			"created_at": token.CreatedAt,
			"expires_at": token.ExpiresAt,
		}
	}

	c.JSON(http.StatusOK, response)
}

// ResendEmailVerification handles POST /api/v1/user/email/resend-verification
func (h *UserHandlers) ResendEmailVerification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Use the auth service to resend verification email
	err := h.authService.ResendVerificationEmail(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to resend verification email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send verification email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent successfully",
	})
}

// GetEmailPreferences handles GET /api/v1/user/email/preferences
func (h *UserHandlers) GetEmailPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// For now, return default preferences
	// In a full implementation, this would query user preferences from database
	h.logger.WithField("user_id", userID).Info("Email preferences retrieved")

	c.JSON(http.StatusOK, gin.H{
		"issues_and_prs":     true,
		"repository_updates": true,
		"security_alerts":    true,
		"mfa_notifications":  true,
		"password_reset":     true,
		"email_verification": true,
		"weekly_digest":      false,
		"marketing_emails":   false,
	})
}

// UpdateEmailPreferences handles PUT /api/v1/user/email/preferences
func (h *UserHandlers) UpdateEmailPreferences(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		IssuesAndPRs      *bool `json:"issues_and_prs,omitempty"`
		RepositoryUpdates *bool `json:"repository_updates,omitempty"`
		SecurityAlerts    *bool `json:"security_alerts,omitempty"`
		MFANotifications  *bool `json:"mfa_notifications,omitempty"`
		PasswordReset     *bool `json:"password_reset,omitempty"`
		EmailVerification *bool `json:"email_verification,omitempty"`
		WeeklyDigest      *bool `json:"weekly_digest,omitempty"`
		MarketingEmails   *bool `json:"marketing_emails,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, just return success with the updated preferences
	// In a full implementation, this would save preferences to database
	h.logger.WithField("user_id", userID).Info("Email preferences updated")

	c.JSON(http.StatusOK, gin.H{
		"message": "Email preferences updated successfully",
		"preferences": gin.H{
			"issues_and_prs":     req.IssuesAndPRs,
			"repository_updates": req.RepositoryUpdates,
			"security_alerts":    req.SecurityAlerts,
			"mfa_notifications":  req.MFANotifications,
			"password_reset":     req.PasswordReset,
			"email_verification": req.EmailVerification,
			"weekly_digest":      req.WeeklyDigest,
			"marketing_emails":   req.MarketingEmails,
		},
	})
}
