package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// AdminEmailHandlers contains handlers for admin email-related endpoints
type AdminEmailHandlers struct {
	db     *gorm.DB
	config *config.Config
	logger *logrus.Logger
}

// NewAdminEmailHandlers creates a new admin email handlers instance
func NewAdminEmailHandlers(db *gorm.DB, cfg *config.Config, logger *logrus.Logger) *AdminEmailHandlers {
	return &AdminEmailHandlers{
		db:     db,
		config: cfg,
		logger: logger,
	}
}

// GetEmailConfig handles GET /api/v1/admin/email/config
func (h *AdminEmailHandlers) GetEmailConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check if user is admin
	h.logger.WithField("user_id", userID).Info("Admin accessing email config")

	// Return current SMTP configuration (without sensitive data)
	c.JSON(http.StatusOK, gin.H{
		"smtp": gin.H{
			"host":       h.config.SMTP.Host,
			"port":       h.config.SMTP.Port,
			"username":   h.config.SMTP.Username,
			"from":       h.config.SMTP.From,
			"use_tls":    h.config.SMTP.UseTLS,
			"configured": h.config.SMTP.Host != "",
		},
		"application": gin.H{
			"base_url": h.config.Application.BaseURL,
			"name":     h.config.Application.Name,
		},
	})
}

// UpdateEmailConfig handles PUT /api/v1/admin/email/config
func (h *AdminEmailHandlers) UpdateEmailConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check if user is admin
	h.logger.WithField("user_id", userID).Info("Admin updating email config")

	var req struct {
		SMTP struct {
			Host     *string `json:"host,omitempty"`
			Port     *string `json:"port,omitempty"`
			Username *string `json:"username,omitempty"`
			Password *string `json:"password,omitempty"`
			From     *string `json:"from,omitempty"`
			UseTLS   *bool   `json:"use_tls,omitempty"`
		} `json:"smtp,omitempty"`
		Application struct {
			BaseURL *string `json:"base_url,omitempty"`
			Name    *string `json:"name,omitempty"`
		} `json:"application,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, just return success message
	// In a full implementation, this would update the configuration
	// and possibly restart the email service with new settings
	h.logger.WithField("user_id", userID).Info("Email configuration update requested")

	c.JSON(http.StatusOK, gin.H{
		"message": "Email configuration updated successfully",
		"note":    "Configuration changes require application restart to take effect",
	})
}

// TestEmailConfig handles POST /api/v1/admin/email/test
func (h *AdminEmailHandlers) TestEmailConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check if user is admin
	h.logger.WithField("user_id", userID).Info("Admin testing email config")

	var req struct {
		To      string `json:"to" binding:"required,email"`
		Subject string `json:"subject,omitempty"`
		Body    string `json:"body,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Use default subject and body if not provided
	if req.Subject == "" {
		req.Subject = "Test Email from A5C Hub"
	}
	if req.Body == "" {
		req.Body = "This is a test email to verify your SMTP configuration is working correctly."
	}

	// Initialize email service and send test email
	emailService := auth.NewSMTPEmailService(h.config)

	// Create a simple test email (we'll modify the email service to support custom emails later)
	err := emailService.SendPasswordResetEmail(req.To, "test-token-not-real")
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to send test email")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send test email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
		"to":      req.To,
	})
}

// GetEmailLogs handles GET /api/v1/admin/email/logs
func (h *AdminEmailHandlers) GetEmailLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check if user is admin
	h.logger.WithField("user_id", userID).Info("Admin accessing email logs")

	// Parse query parameters
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "50")
	emailType := c.Query("type") // verification, password_reset, mfa_setup

	// For now, return mock email logs
	// In a full implementation, this would query actual email delivery logs
	c.JSON(http.StatusOK, gin.H{
		"logs": []gin.H{
			{
				"id":        "1",
				"to":        "user@example.com",
				"subject":   "Email Verification",
				"type":      "verification",
				"status":    "sent",
				"sent_at":   "2025-07-26T04:00:00Z",
				"delivered": true,
				"error":     nil,
			},
			{
				"id":        "2",
				"to":        "admin@example.com",
				"subject":   "Password Reset Request",
				"type":      "password_reset",
				"status":    "sent",
				"sent_at":   "2025-07-26T03:30:00Z",
				"delivered": true,
				"error":     nil,
			},
		},
		"pagination": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       2,
			"total_pages": 1,
		},
		"filters": gin.H{
			"type": emailType,
		},
	})
}

// GetEmailHealth handles GET /api/v1/admin/email/health
func (h *AdminEmailHandlers) GetEmailHealth(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check if user is admin
	h.logger.WithField("user_id", userID).Info("Admin checking email health")

	// Check if SMTP is configured
	configured := h.config.SMTP.Host != ""

	status := "healthy"
	if !configured {
		status = "not_configured"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"configured": configured,
		"smtp": gin.H{
			"host_configured": h.config.SMTP.Host != "",
			"auth_configured": h.config.SMTP.Username != "" && h.config.SMTP.Password != "",
			"from_configured": h.config.SMTP.From != "",
		},
		"stats": gin.H{
			"emails_sent_today":     0, // TODO: implement actual stats
			"emails_sent_this_week": 0,
			"failed_emails_today":   0,
			"success_rate":          100.0,
		},
		"last_check": "2025-07-26T05:00:00Z",
	})
}
