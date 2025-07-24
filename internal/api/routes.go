package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(router *gin.Engine, database *db.Database, logger *logrus.Logger) {
	cfg, _ := config.Load()
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// Initialize authentication services
	authService := auth.NewAuthService(database.DB, jwtManager, cfg)
	oauthConfig := auth.OAuthConfig{
		GitHub: auth.GitHubConfig{
			ClientID:     "your-github-client-id",
			ClientSecret: "your-github-client-secret",
			RedirectURL:  "http://localhost:8080/api/v1/auth/oauth/github/callback",
		},
	}
	oauthService := auth.NewOAuthService(database.DB, oauthConfig, authService)
	mfaService := auth.NewMFAService(database.DB)
	authHandlers := NewAuthHandlers(authService, oauthService, mfaService)

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
			// Basic authentication
			authGroup.POST("/login", authHandlers.Login)
			authGroup.POST("/register", authHandlers.Register)
			authGroup.POST("/refresh", authHandlers.RefreshToken)
			authGroup.POST("/forgot-password", authHandlers.ForgotPassword)
			authGroup.POST("/reset-password", authHandlers.ResetPassword)
			authGroup.GET("/verify-email", authHandlers.VerifyEmail)
			
			// OAuth endpoints
			oauth := authGroup.Group("/oauth")
			{
				oauth.GET("/:provider", authHandlers.OAuthRedirect)
				oauth.GET("/:provider/callback", authHandlers.OAuthCallback)
			}
			
			// Protected auth endpoints
			protected := authGroup.Group("/")
			protected.Use(middleware.AuthMiddleware(jwtManager))
			{
				protected.POST("/logout", authHandlers.Logout)
				
				// MFA endpoints
				mfa := protected.Group("/mfa")
				{
					mfa.POST("/setup", authHandlers.SetupMFA)
					mfa.POST("/verify", authHandlers.VerifyMFA)
					mfa.POST("/backup-codes", authHandlers.RegenerateBackupCodes)
					mfa.DELETE("/disable", authHandlers.DisableMFA)
				}
			}
		}

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userID, _ := c.Get("user_id")
				username, _ := c.Get("username")
				email, _ := c.Get("email")
				
				c.JSON(http.StatusOK, gin.H{
					"user_id":  userID,
					"username": username,
					"email":    email,
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