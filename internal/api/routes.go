package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(router *gin.Engine, database *db.Database, logger *logrus.Logger) {
	cfg, _ := config.Load()
	jwtManager := auth.NewJWTManager(cfg.JWT)
	authService := auth.NewAuthService(database, cfg)

	router.GET("/health", func(c *gin.Context) {
		if err := database.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z",
			"version":   "1.0.0",
		})
	})

	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", func(c *gin.Context) {
				var req auth.LoginRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := authService.Login(&req)
				if err != nil {
					switch err {
					case auth.ErrInvalidCredentials:
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
					case auth.ErrAccountLocked:
						c.JSON(http.StatusForbidden, gin.H{"error": "Account is locked"})
					default:
						logger.WithError(err).Error("Login failed")
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    response,
				})
			})

			authGroup.POST("/register", func(c *gin.Context) {
				var req auth.RegisterRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				response, err := authService.Register(&req)
				if err != nil {
					switch err {
					case auth.ErrUserExists:
						c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
					default:
						logger.WithError(err).Error("Registration failed")
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
					return
				}

				c.JSON(http.StatusCreated, gin.H{
					"success": true,
					"data":    response,
				})
			})

			authGroup.POST("/logout", func(c *gin.Context) {
				// For JWT-based auth, logout is client-side (remove token)
				// In the future, we could implement token blacklisting
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Logged out successfully",
				})
			})

			authGroup.POST("/forgot-password", func(c *gin.Context) {
				var req auth.PasswordResetRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				err := authService.InitiatePasswordReset(req.Email)
				if err != nil {
					logger.WithError(err).Error("Password reset initiation failed")
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					return
				}

				// Always return success for security (don't reveal if email exists)
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "If an account with that email exists, a password reset link has been sent",
				})
			})

			authGroup.POST("/reset-password", func(c *gin.Context) {
				var req auth.PasswordResetConfirmRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				err := authService.ResetPassword(req.Token, req.Password)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Password reset successfully",
				})
			})

			authGroup.POST("/verify-email", func(c *gin.Context) {
				token := c.Query("token")
				if token == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
					return
				}

				err := authService.VerifyEmail(token)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Email verified successfully",
				})
			})

			// OAuth endpoints
			authGroup.GET("/oauth/:provider", func(c *gin.Context) {
				provider := c.Param("provider")
				redirectURI := c.Query("redirect_uri")
				if redirectURI == "" {
					redirectURI = "http://localhost:3000/auth/callback/" + provider
				}

				authURL, state, err := authService.InitiateOAuth(provider, redirectURI)
				if err != nil {
					switch err {
					case auth.ErrOAuthProviderNotConfigured:
						c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth provider not configured"})
					default:
						logger.WithError(err).Error("OAuth initiation failed")
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success":  true,
					"auth_url": authURL,
					"state":    state,
				})
			})

			authGroup.GET("/oauth/:provider/callback", func(c *gin.Context) {
				provider := c.Param("provider")
				code := c.Query("code")
				state := c.Query("state")
				redirectURI := c.Query("redirect_uri")
				
				if code == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
					return
				}

				if redirectURI == "" {
					redirectURI = "http://localhost:3000/auth/callback/" + provider
				}

				response, err := authService.CompleteOAuth(provider, code, state, redirectURI)
				if err != nil {
					switch err {
					case auth.ErrOAuthProviderNotConfigured:
						c.JSON(http.StatusBadRequest, gin.H{"error": "OAuth provider not configured"})
					case auth.ErrInvalidOAuthState:
						c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OAuth state"})
					case auth.ErrOAuthCodeExchange:
						c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to exchange authorization code"})
					case auth.ErrOAuthUserInfo:
						c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve user information"})
					default:
						logger.WithError(err).Error("OAuth completion failed")
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
					}
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    response,
				})
			})
		}

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
					return
				}

				user, err := authService.GetUserByID(userID.(uuid.UUID))
				if err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data":    user,
				})
			})

			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusNotImplemented, gin.H{"message": "Admin users endpoint - to be implemented"})
				})
			}

			repos := protected.Group("/repositories")
			{
				repos.GET("/", func(c *gin.Context) {
					c.JSON(http.StatusNotImplemented, gin.H{"message": "List repositories endpoint - to be implemented"})
				})
				repos.POST("/", func(c *gin.Context) {
					c.JSON(http.StatusNotImplemented, gin.H{"message": "Create repository endpoint - to be implemented"})
				})
				repos.GET("/:owner/:repo", func(c *gin.Context) {
					c.JSON(http.StatusNotImplemented, gin.H{"message": "Get repository endpoint - to be implemented"})
				})
			}
		}
	}
}