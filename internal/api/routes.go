package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/middleware"
	"github.com/a5c-ai/hub/internal/services"
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

	// Initialize Git services
	gitService := git.NewGitService(logger)
	repoBasePath := cfg.Storage.RepositoryPath
	if repoBasePath == "" {
		repoBasePath = "/var/lib/hub/repositories"
	}
	
	repositoryService := services.NewRepositoryService(database.DB, gitService, logger, repoBasePath)
	branchService := services.NewBranchService(database.DB, gitService, repositoryService, logger)

	// Initialize handlers
	repoHandlers := NewRepositoryHandlers(repositoryService, branchService, logger)
	gitHandlers := NewGitHandlers(repositoryService, logger)

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

	// Git HTTP protocol endpoints (no authentication required for public repos)
	git := router.Group("/")
	git.Use(gitHandlers.GitMiddleware())
	{
		git.GET("/:owner/:repo.git/info/refs", gitHandlers.InfoRefs)
		git.POST("/:owner/:repo.git/git-upload-pack", gitHandlers.UploadPack)
		git.POST("/:owner/:repo.git/git-receive-pack", gitHandlers.ReceivePack)
	}

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

		// Public repository endpoints (for public repos)
		v1.GET("/repositories", repoHandlers.ListRepositories)
		v1.GET("/repositories/:owner/:repo", repoHandlers.GetRepository)
		v1.GET("/repositories/:owner/:repo/branches", repoHandlers.GetBranches)
		v1.GET("/repositories/:owner/:repo/branches/:branch", repoHandlers.GetBranch)

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

			// Protected repository endpoints
			repos := protected.Group("/repositories")
			{
				repos.POST("/", repoHandlers.CreateRepository)
				repos.PATCH("/:owner/:repo", repoHandlers.UpdateRepository)
				repos.DELETE("/:owner/:repo", repoHandlers.DeleteRepository)
				
				// Branch operations
				repos.POST("/:owner/:repo/branches", repoHandlers.CreateBranch)
				repos.DELETE("/:owner/:repo/branches/:branch", repoHandlers.DeleteBranch)
			}
		}
	}
}