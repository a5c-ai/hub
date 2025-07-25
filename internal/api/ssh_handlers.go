package api

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// SSHKeyHandlers handles SSH key related operations
type SSHKeyHandlers struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewSSHKeyHandlers creates a new SSH key handlers instance
func NewSSHKeyHandlers(db *gorm.DB, logger *logrus.Logger) *SSHKeyHandlers {
	return &SSHKeyHandlers{
		db:     db,
		logger: logger,
	}
}

// CreateSSHKeyRequest represents a request to create an SSH key
type CreateSSHKeyRequest struct {
	Title   string `json:"title" binding:"required"`
	KeyData string `json:"key_data" binding:"required"`
}

// SSHKeyResponse represents an SSH key response
type SSHKeyResponse struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Fingerprint string     `json:"fingerprint"`
	KeyType     string     `json:"key_type"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ListSSHKeys handles GET /api/v1/user/keys
func (h *SSHKeyHandlers) ListSSHKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var sshKeys []models.SSHKey
	if err := h.db.Where("user_id = ?", uid).Order("created_at DESC").Find(&sshKeys).Error; err != nil {
		h.logger.WithError(err).Error("Failed to fetch SSH keys")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SSH keys"})
		return
	}

	var response []SSHKeyResponse
	for _, key := range sshKeys {
		keyType := h.getKeyType(key.KeyData)
		response = append(response, SSHKeyResponse{
			ID:          key.ID,
			Title:       key.Title,
			Fingerprint: key.Fingerprint,
			KeyType:     keyType,
			LastUsedAt:  key.LastUsedAt,
			CreatedAt:   key.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// CreateSSHKey handles POST /api/v1/user/keys
func (h *SSHKeyHandlers) CreateSSHKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate SSH key format
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(req.KeyData))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSH key format"})
		return
	}

	// Generate fingerprint
	fingerprint := h.generateFingerprint(publicKey)

	// Check if fingerprint already exists
	var existingKey models.SSHKey
	if err := h.db.Where("fingerprint = ?", fingerprint).First(&existingKey).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "SSH key already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		h.logger.WithError(err).Error("Failed to check existing SSH key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate SSH key"})
		return
	}

	// Create SSH key
	sshKey := models.SSHKey{
		UserID:      uid,
		Title:       req.Title,
		KeyData:     strings.TrimSpace(req.KeyData),
		Fingerprint: fingerprint,
	}

	if err := h.db.Create(&sshKey).Error; err != nil {
		h.logger.WithError(err).Error("Failed to create SSH key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SSH key"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id":     uid,
		"key_id":      sshKey.ID,
		"fingerprint": fingerprint,
	}).Info("SSH key created")

	response := SSHKeyResponse{
		ID:          sshKey.ID,
		Title:       sshKey.Title,
		Fingerprint: sshKey.Fingerprint,
		KeyType:     h.getKeyType(sshKey.KeyData),
		LastUsedAt:  sshKey.LastUsedAt,
		CreatedAt:   sshKey.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetSSHKey handles GET /api/v1/user/keys/:id
func (h *SSHKeyHandlers) GetSSHKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	var sshKey models.SSHKey
	if err := h.db.Where("id = ? AND user_id = ?", keyID, uid).First(&sshKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		} else {
			h.logger.WithError(err).Error("Failed to fetch SSH key")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SSH key"})
		}
		return
	}

	response := SSHKeyResponse{
		ID:          sshKey.ID,
		Title:       sshKey.Title,
		Fingerprint: sshKey.Fingerprint,
		KeyType:     h.getKeyType(sshKey.KeyData),
		LastUsedAt:  sshKey.LastUsedAt,
		CreatedAt:   sshKey.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSSHKey handles DELETE /api/v1/user/keys/:id
func (h *SSHKeyHandlers) DeleteSSHKey(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	keyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", keyID, uid).Delete(&models.SSHKey{})
	if result.Error != nil {
		h.logger.WithError(result.Error).Error("Failed to delete SSH key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SSH key"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"user_id": uid,
		"key_id":  keyID,
	}).Info("SSH key deleted")

	c.JSON(http.StatusNoContent, nil)
}

// generateFingerprint generates an SSH key fingerprint
func (h *SSHKeyHandlers) generateFingerprint(publicKey ssh.PublicKey) string {
	hash := sha256.Sum256(publicKey.Marshal())
	return fmt.Sprintf("SHA256:%s", base64.StdEncoding.EncodeToString(hash[:]))
}

// getKeyType extracts the key type from SSH key data
func (h *SSHKeyHandlers) getKeyType(keyData string) string {
	parts := strings.Fields(keyData)
	if len(parts) < 2 {
		return "unknown"
	}
	return parts[0]
}

// generateMD5Fingerprint generates MD5 fingerprint (legacy format)
func (h *SSHKeyHandlers) generateMD5Fingerprint(publicKey ssh.PublicKey) string {
	hash := md5.Sum(publicKey.Marshal())
	var fingerprint strings.Builder
	for i, b := range hash {
		if i > 0 {
			fingerprint.WriteString(":")
		}
		fingerprint.WriteString(fmt.Sprintf("%02x", b))
	}
	return fingerprint.String()
}
