package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/controllers"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/git"
	"github.com/a5c-ai/hub/internal/middleware"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(router *gin.Engine, database *db.Database, logger *logrus.Logger) {
	cfg, _ := config.Load()
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// Initialize authentication services
	authService := auth.NewAuthService(database.DB, jwtManager, cfg)
	oauthService := auth.NewOAuthService(database.DB, jwtManager, cfg, authService)
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
	pullRequestService := services.NewPullRequestService(database.DB, gitService, repositoryService, logger, repoBasePath)

	// Initialize organization services
	activityService := services.NewActivityService(database.DB)
	orgService := services.NewOrganizationService(database.DB, activityService)
	memberService := services.NewMembershipService(database.DB, activityService)
	invitationService := services.NewInvitationService(database.DB, activityService)
	teamService := services.NewTeamService(database.DB, activityService)
	teamMembershipService := services.NewTeamMembershipService(database.DB, activityService)
	permissionService := services.NewPermissionService(database.DB, activityService)

	// Initialize Elasticsearch service
	elasticsearchService, err := services.NewElasticsearchService(&cfg.Elasticsearch, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Elasticsearch service")
	}

	// Initialize search service
	searchService := services.NewSearchService(database.DB, elasticsearchService, logger)

	// Initialize analytics service
	analyticsService := services.NewAnalyticsService(database.DB, logger)

	// Initialize notification service for real-time push
	notificationService := services.NewNotificationService()

	// Initialize handlers
	repoHandlers := NewRepositoryHandlers(repositoryService, branchService, gitService, logger, database.DB)
	gitHandlers := NewGitHandlers(repositoryService, logger)
	prHandlers := NewPullRequestHandlers(pullRequestService, logger)
	searchHandlers := NewSearchHandlers(searchService, logger)

	userHandlers := NewUserHandlers(authService, database.DB, cfg, logger, notificationService)
	adminEmailHandlers := NewAdminEmailHandlers(database.DB, cfg, logger)
	activityHandlers := NewActivityHandlers(repositoryService, activityService, database.DB, logger)
	// Initialize webhook and deploy key services for hooks handlers
	webhookDeliveryService := services.NewWebhookDeliveryService(database.DB, logger)
	deployKeyService := services.NewDeployKeyService(database.DB, logger)
	hooksHandlers := NewHooksHandlers(repositoryService, webhookDeliveryService, deployKeyService, logger)
	branchProtectionHandlers := NewBranchProtectionHandlers(repositoryService, branchService, logger)
	analyticsHandlers := NewAnalyticsHandlers(analyticsService, logger, database.DB)
	sshKeyHandlers := NewSSHKeyHandlers(database.DB, logger)
	adminHandlers := NewAdminHandlers(authService, database.DB, logger)

	// Initialize import/export handlers
	importHandlers := NewImportHandlers(database)
	exportHandlers := NewExportHandlers(database)

	orgController := controllers.NewOrganizationController(orgService, memberService, invitationService, activityService)
	teamController := controllers.NewTeamController(teamService, teamMembershipService, permissionService)

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
		git.GET("/:owner/:repo/info/refs", gitHandlers.InfoRefs)
		git.POST("/:owner/:repo/git-upload-pack", gitHandlers.UploadPack)
		git.POST("/:owner/:repo/git-receive-pack", gitHandlers.ReceivePack)
	}

	v1 := router.Group("/api/v1")
	{
		// Git LFS endpoints (batch API, upload, download, verify)
		lfsHandlers, err := NewLFSHandlers(cfg.LFS, repoBasePath)
		if err != nil {
			logger.WithError(err).Fatal("failed to initialize Git LFS handlers")
		}
		lfs := v1.Group("/git-lfs")
		{
			lfs.POST("/objects/batch", lfsHandlers.Batch)
			lfs.POST("/objects/:oid", lfsHandlers.Upload)
			lfs.GET("/objects/:oid", lfsHandlers.Download)
			lfs.HEAD("/objects/:oid", lfsHandlers.Verify)
		}
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
					mfa.POST("/disable", authHandlers.DisableMFA)
					mfa.POST("/regenerate-codes", authHandlers.RegenerateBackupCodes)
				}
			}
		}

		// Public repository endpoints (for public repos)
		v1.GET("/repositories", repoHandlers.ListRepositories)
		v1.GET("/repositories/:owner/:repo", repoHandlers.GetRepository)
		v1.GET("/repositories/:owner/:repo/branches", repoHandlers.GetBranches)
		v1.GET("/repositories/:owner/:repo/branches/:branch", repoHandlers.GetBranch)

		// Git content endpoints (public access)
		v1.GET("/repositories/:owner/:repo/commits", repoHandlers.GetCommits)
		v1.GET("/repositories/:owner/:repo/commits/:sha", repoHandlers.GetCommit)
		v1.GET("/repositories/:owner/:repo/contents/*path", repoHandlers.GetTree)
		v1.GET("/repositories/:owner/:repo/info", repoHandlers.GetRepositoryInfo)

		// Public search endpoints (for public content)
		v1.GET("/search", searchHandlers.GlobalSearch)

		// Public user profile endpoints
		v1.GET("/users/:username", userHandlers.GetUserProfile)
		v1.GET("/users/:username/repositories", userHandlers.GetUserRepositories)
		v1.GET("/users/:username/organizations", userHandlers.GetUserOrganizations)
		v1.GET("/users/:username/analytics/public", analyticsHandlers.GetPublicUserAnalytics)

		// Public invitation acceptance endpoint
		v1.POST("/invitations/accept", orgController.AcceptInvitation)

		// Webhook endpoints (no authentication required for system-level webhooks)

		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// Current user profile endpoints
			protected.GET("/user", userHandlers.GetCurrentUserProfile)
			protected.PATCH("/user", userHandlers.UpdateUserProfile)

			// User activity and notifications
			protected.GET("/user/activity", userHandlers.GetUserActivity)
			protected.GET("/notifications", userHandlers.GetNotifications)
			protected.PATCH("/notifications", userHandlers.MarkNotificationsAsRead)
			// Real-time notifications via WebSocket
			protected.GET("/notifications/subscribe", userHandlers.SubscribeNotifications)

			// User email endpoints
			emailGroup := protected.Group("/user/email")
			{
				emailGroup.GET("/verification-status", userHandlers.GetEmailVerificationStatus)
				emailGroup.POST("/resend-verification", userHandlers.ResendEmailVerification)
				emailGroup.GET("/preferences", userHandlers.GetEmailPreferences)
				emailGroup.PUT("/preferences", userHandlers.UpdateEmailPreferences)
			}

			// Legacy profile endpoint for backward compatibility
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

			// User's organizations
			protected.GET("/user/organizations", orgController.GetUserOrganizations)

			// User analytics endpoints
			protected.GET("/user/analytics/activity", analyticsHandlers.GetUserAnalytics)
			protected.GET("/user/analytics/contributions", analyticsHandlers.GetUserContributions)
			protected.GET("/user/analytics/repositories", analyticsHandlers.GetUserRepositories)

			// SSH Keys management
			protected.GET("/user/keys", sshKeyHandlers.ListSSHKeys)
			protected.POST("/user/keys", sshKeyHandlers.CreateSSHKey)
			protected.GET("/user/keys/:id", sshKeyHandlers.GetSSHKey)
			protected.DELETE("/user/keys/:id", sshKeyHandlers.DeleteSSHKey)

			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				// Admin user management endpoints
				admin.GET("/users", adminHandlers.ListUsers)
				admin.POST("/users", adminHandlers.CreateUser)
				admin.GET("/users/stats", adminHandlers.GetUserStats)
				admin.GET("/users/:id", adminHandlers.GetUser)
				admin.PATCH("/users/:id", adminHandlers.UpdateUser)
				admin.DELETE("/users/:id", adminHandlers.DeleteUser)
				admin.POST("/users/:id/enable", adminHandlers.EnableUser)
				admin.POST("/users/:id/disable", adminHandlers.DisableUser)
				admin.PATCH("/users/:id/role", adminHandlers.SetUserRole)

				// Admin analytics endpoints
				admin.GET("/analytics/platform", analyticsHandlers.GetPlatformAnalytics)
				admin.GET("/analytics/usage", analyticsHandlers.GetUsageAnalytics)
				admin.GET("/analytics/performance", analyticsHandlers.GetPerformanceAnalytics)
				admin.GET("/analytics/costs", analyticsHandlers.GetCostAnalytics)
				admin.GET("/analytics/export", analyticsHandlers.ExportAnalytics)

				// Admin email management endpoints
				adminEmail := admin.Group("/email")
				{
					adminEmail.GET("/config", adminEmailHandlers.GetEmailConfig)
					adminEmail.PUT("/config", adminEmailHandlers.UpdateEmailConfig)
					adminEmail.POST("/test", adminEmailHandlers.TestEmailConfig)
					adminEmail.GET("/logs", adminEmailHandlers.GetEmailLogs)
					adminEmail.GET("/health", adminEmailHandlers.GetEmailHealth)
				}

				// Storage admin endpoints

			}

			// Protected repository endpoints
			// Repository import/export endpoints
			protected.POST("/repositories/import", importHandlers.InitiateImport)
			protected.GET("/repositories/import/:job_id", importHandlers.GetImportStatus)
			protected.POST("/repositories/:owner/:repo/export", exportHandlers.InitiateExport)
			protected.GET("/repositories/:owner/:repo/export/:job_id", exportHandlers.GetExportStatus)

			// Repository creation endpoint (without group to avoid trailing slash issues)
			protected.POST("/repositories", repoHandlers.CreateRepository)

			repos := protected.Group("/repositories")
			{
				repos.PATCH("/:owner/:repo", repoHandlers.UpdateRepository)
				repos.DELETE("/:owner/:repo", repoHandlers.DeleteRepository)

				// Branch operations
				repos.POST("/:owner/:repo/branches", repoHandlers.CreateBranch)
				repos.DELETE("/:owner/:repo/branches/:branch", repoHandlers.DeleteBranch)

				// File operations
				repos.POST("/:owner/:repo/contents/*path", repoHandlers.CreateFile)
				repos.PUT("/:owner/:repo/contents/*path", repoHandlers.UpdateFile)
				repos.DELETE("/:owner/:repo/contents/*path", repoHandlers.DeleteFile)

				// Repository information and statistics
				repos.GET("/:owner/:repo/stats", repoHandlers.GetRepositoryStats)
				repos.GET("/:owner/:repo/languages", repoHandlers.GetRepositoryLanguages)
				repos.GET("/:owner/:repo/tags", repoHandlers.GetRepositoryTags)
				repos.GET("/:owner/:repo/contributors", activityHandlers.GetRepositoryContributors)
				repos.GET("/:owner/:repo/activity", activityHandlers.GetRepositoryActivity)

				// Branch comparison
				repos.GET("/:owner/:repo/compare/:base/:head", repoHandlers.CompareBranches)
				repos.GET("/:owner/:repo/compare/:base/head", repoHandlers.GetMergeBase)

				// Branch protection
				repos.GET("/:owner/:repo/branches/:branch/protection", branchProtectionHandlers.GetBranchProtection)
				repos.PUT("/:owner/:repo/branches/:branch/protection", branchProtectionHandlers.UpdateBranchProtection)
				repos.DELETE("/:owner/:repo/branches/:branch/protection", branchProtectionHandlers.DeleteBranchProtection)
				repos.GET("/:owner/:repo/branches/:branch/protection/required_status_checks", branchProtectionHandlers.GetRequiredStatusChecks)
				repos.PATCH("/:owner/:repo/branches/:branch/protection/required_status_checks", branchProtectionHandlers.UpdateRequiredStatusChecks)
				repos.DELETE("/:owner/:repo/branches/:branch/protection/required_status_checks", branchProtectionHandlers.DeleteRequiredStatusChecks)
				repos.GET("/:owner/:repo/branches/:branch/protection/required_pull_request_reviews", branchProtectionHandlers.GetRequiredPullRequestReviews)
				repos.PATCH("/:owner/:repo/branches/:branch/protection/required_pull_request_reviews", branchProtectionHandlers.UpdateRequiredPullRequestReviews)
				repos.DELETE("/:owner/:repo/branches/:branch/protection/required_pull_request_reviews", branchProtectionHandlers.DeleteRequiredPullRequestReviews)

				// Webhooks
				repos.GET("/:owner/:repo/hooks", hooksHandlers.ListWebhooks)
				repos.POST("/:owner/:repo/hooks", hooksHandlers.CreateWebhook)
				repos.GET("/:owner/:repo/hooks/:hook_id", hooksHandlers.GetWebhook)
				repos.PATCH("/:owner/:repo/hooks/:hook_id", hooksHandlers.UpdateWebhook)
				repos.DELETE("/:owner/:repo/hooks/:hook_id", hooksHandlers.DeleteWebhook)
				repos.POST("/:owner/:repo/hooks/:hook_id/pings", hooksHandlers.PingWebhook)

				// Deploy keys
				repos.GET("/:owner/:repo/keys", hooksHandlers.ListDeployKeys)
				repos.POST("/:owner/:repo/keys", hooksHandlers.CreateDeployKey)
				repos.GET("/:owner/:repo/keys/:key_id", hooksHandlers.GetDeployKey)
				repos.DELETE("/:owner/:repo/keys/:key_id", hooksHandlers.DeleteDeployKey)

				// Repository subscription (watching)
				repos.GET("/:owner/:repo/subscription", activityHandlers.GetRepositorySubscription)
				repos.PUT("/:owner/:repo/subscription", activityHandlers.WatchRepository)
				repos.DELETE("/:owner/:repo/subscription", activityHandlers.UnwatchRepository)

				// Repository starring
				repos.GET("/:owner/:repo/star", repoHandlers.CheckStarred)
				repos.PUT("/:owner/:repo/star", repoHandlers.StarRepository)
				repos.DELETE("/:owner/:repo/star", repoHandlers.UnstarRepository)

				// Repository forking
				repos.POST("/:owner/:repo/fork", repoHandlers.ForkRepository)

				// Repository settings read/write in dedicated branch
				repos.GET("/:owner/:repo/settings", repoHandlers.GetRepositorySettings)
				repos.PUT("/:owner/:repo/settings", repoHandlers.UpdateRepositorySettings)

				// Repository-specific search

				// Pull request operations
				repos.GET("/:owner/:repo/pulls", prHandlers.ListPullRequests)
				repos.POST("/:owner/:repo/pulls", prHandlers.CreatePullRequest)
				repos.GET("/:owner/:repo/pulls/:number", prHandlers.GetPullRequest)
				repos.PATCH("/:owner/:repo/pulls/:number", prHandlers.UpdatePullRequest)
				repos.PUT("/:owner/:repo/pulls/:number/merge", prHandlers.MergePullRequest)

				// Repository analytics endpoints (require authentication)
				repos.GET("/:owner/:repo/analytics", analyticsHandlers.GetRepositoryAnalytics)
				repos.GET("/:owner/:repo/analytics/code-stats", analyticsHandlers.GetRepositoryCodeStats)
				repos.GET("/:owner/:repo/analytics/contributors", analyticsHandlers.GetRepositoryContributors)
				repos.GET("/:owner/:repo/analytics/activity", analyticsHandlers.GetRepositoryActivity)
				repos.GET("/:owner/:repo/analytics/performance", analyticsHandlers.GetRepositoryPerformance)
				repos.GET("/:owner/:repo/analytics/issues", analyticsHandlers.GetRepositoryIssues)
				repos.GET("/:owner/:repo/analytics/pulls", analyticsHandlers.GetRepositoryPulls)
			}

			// Admin-only operations
			adminRepos := protected.Group("/repositories")
			adminRepos.Use(middleware.AdminMiddleware())
			{

			}

			// Organization management endpoints
			orgs := protected.Group("/organizations")
			{
				orgs.GET("/", orgController.ListOrganizations)
				orgs.POST("/", orgController.CreateOrganization)
				orgs.GET("/:org", orgController.GetOrganization)
				orgs.PATCH("/:org", orgController.UpdateOrganization)
				orgs.DELETE("/:org", orgController.DeleteOrganization)

				// Organization members
				orgs.GET("/:org/members", orgController.GetMembers)
				orgs.GET("/:org/members/:username", orgController.GetMember)
				orgs.PUT("/:org/members/:username", orgController.AddMember)
				orgs.DELETE("/:org/members/:username", orgController.RemoveMember)
				orgs.PATCH("/:org/members/:username", orgController.UpdateMemberRole)

				// Public/Private membership
				orgs.PUT("/:org/public_members/:username", orgController.SetMemberPublic)
				orgs.DELETE("/:org/public_members/:username", orgController.SetMemberPrivate)

				// Organization invitations
				orgs.GET("/:org/invitations", orgController.GetInvitations)
				orgs.POST("/:org/invitations", orgController.CreateInvitation)
				orgs.DELETE("/:org/invitations/:invitation_id", orgController.CancelInvitation)

				// Organization activity
				orgs.GET("/:org/activity", orgController.GetActivity)

				// Organization teams
				orgs.GET("/:org/teams", teamController.ListTeams)
				orgs.POST("/:org/teams", teamController.CreateTeam)
				orgs.GET("/:org/teams/hierarchy", teamController.GetTeamHierarchy)
				orgs.GET("/:org/teams/:team", teamController.GetTeam)
				orgs.PATCH("/:org/teams/:team", teamController.UpdateTeam)
				orgs.DELETE("/:org/teams/:team", teamController.DeleteTeam)

				// Team members
				orgs.GET("/:org/teams/:team/members", teamController.GetTeamMembers)
				orgs.PUT("/:org/teams/:team/members/:username", teamController.AddTeamMember)
				orgs.DELETE("/:org/teams/:team/members/:username", teamController.RemoveTeamMember)
				orgs.PATCH("/:org/teams/:team/members/:username", teamController.UpdateTeamMemberRole)

				// Team repositories
				orgs.GET("/:org/teams/:team/repositories", teamController.GetTeamRepositories)
				orgs.PUT("/:org/teams/:team/repositories/:repo", teamController.AddTeamRepository)
				orgs.DELETE("/:org/teams/:team/repositories/:repo", teamController.RemoveTeamRepository)

				// User teams in organization
				orgs.GET("/:org/members/:username/teams", teamController.GetUserTeams)

				// Organization analytics endpoints
				orgs.GET("/:org/analytics/overview", analyticsHandlers.GetOrganizationAnalytics)
				orgs.GET("/:org/analytics/members", analyticsHandlers.GetOrganizationMembers)
				orgs.GET("/:org/analytics/repositories", analyticsHandlers.GetOrganizationRepositories)
				orgs.GET("/:org/analytics/teams", analyticsHandlers.GetOrganizationTeams)
				orgs.GET("/:org/analytics/security", analyticsHandlers.GetOrganizationSecurity)
			}
		}
	}
}
