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

	// Initialize issue-related services
	issueService := services.NewIssueService(database.DB, logger)
	commentService := services.NewCommentService(database.DB, logger)
	labelService := services.NewLabelService(database.DB, logger)
	milestoneService := services.NewMilestoneService(database.DB, logger)

	// Initialize organization services
	activityService := services.NewActivityService(database.DB)
	orgService := services.NewOrganizationService(database.DB, activityService)
	memberService := services.NewMembershipService(database.DB, activityService)
	invitationService := services.NewInvitationService(database.DB, activityService)
	teamService := services.NewTeamService(database.DB, activityService)
	teamMembershipService := services.NewTeamMembershipService(database.DB, activityService)
	permissionService := services.NewPermissionService(database.DB, activityService)

	// Initialize search service
	searchService := services.NewSearchService(database.DB)

	// Initialize handlers
	repoHandlers := NewRepositoryHandlers(repositoryService, branchService, gitService, logger)
	gitHandlers := NewGitHandlers(repositoryService, logger)
	prHandlers := NewPullRequestHandlers(pullRequestService, logger)
	searchHandlers := NewSearchHandlers(searchService, logger)
	issueHandlers := NewIssueHandlers(issueService, commentService, labelService, milestoneService, repositoryService, logger)
	commentHandlers := NewCommentHandlers(commentService, issueService, logger)
	labelHandlers := NewLabelHandlers(labelService, repositoryService, logger)
	milestoneHandlers := NewMilestoneHandlers(milestoneService, repositoryService, logger)
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

		// Public issue endpoints (for public repos)
		v1.GET("/repositories/:owner/:repo/issues", issueHandlers.ListIssues)
		v1.GET("/repositories/:owner/:repo/issues/search", issueHandlers.SearchIssues)
		v1.GET("/repositories/:owner/:repo/issues/:number", issueHandlers.GetIssue)
		v1.GET("/repositories/:owner/:repo/issues/:number/comments", commentHandlers.ListComments)
		v1.GET("/repositories/:owner/:repo/issues/:number/comments/:comment_id", commentHandlers.GetComment)
		v1.GET("/repositories/:owner/:repo/labels", labelHandlers.ListLabels)
		v1.GET("/repositories/:owner/:repo/labels/:name", labelHandlers.GetLabel)
		v1.GET("/repositories/:owner/:repo/milestones", milestoneHandlers.ListMilestones)
		v1.GET("/repositories/:owner/:repo/milestones/:number", milestoneHandlers.GetMilestone)

		// Public search endpoints (for public content)
		v1.GET("/search", searchHandlers.GlobalSearch)
		v1.GET("/search/repositories", searchHandlers.SearchRepositories)
		v1.GET("/search/issues", searchHandlers.SearchIssues)
		v1.GET("/search/users", searchHandlers.SearchUsers)
		v1.GET("/search/commits", searchHandlers.SearchCommits)

		// Public invitation acceptance endpoint
		v1.POST("/invitations/accept", orgController.AcceptInvitation)

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

			// User's organizations
			protected.GET("/user/organizations", orgController.GetUserOrganizations)

			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.GET("/users", func(c *gin.Context) {
					c.JSON(http.StatusNotImplemented, gin.H{"message": "Admin users endpoint - to be implemented"})
				})
			}

			// Protected repository endpoints  
			// Repository creation endpoint (without group to avoid trailing slash issues)
			protected.POST("/repositories", repoHandlers.CreateRepository)
			
			repos := protected.Group("/repositories")
			{
				repos.PATCH("/:owner/:repo", repoHandlers.UpdateRepository)
				repos.DELETE("/:owner/:repo", repoHandlers.DeleteRepository)

				// Branch operations
				repos.POST("/:owner/:repo/branches", repoHandlers.CreateBranch)
				repos.DELETE("/:owner/:repo/branches/:branch", repoHandlers.DeleteBranch)

				// Repository-specific search
				repos.GET("/:owner/:repo/search", searchHandlers.SearchInRepository)

				// Pull request operations
				repos.GET("/:owner/:repo/pulls", prHandlers.ListPullRequests)
				repos.POST("/:owner/:repo/pulls", prHandlers.CreatePullRequest)
				repos.GET("/:owner/:repo/pulls/:number", prHandlers.GetPullRequest)
				repos.PATCH("/:owner/:repo/pulls/:number", prHandlers.UpdatePullRequest)
				repos.PUT("/:owner/:repo/pulls/:number/merge", prHandlers.MergePullRequest)
				repos.GET("/:owner/:repo/pulls/:number/files", prHandlers.GetPullRequestFiles)

				// Review operations
				repos.POST("/:owner/:repo/pulls/:number/reviews", prHandlers.CreateReview)
				repos.GET("/:owner/:repo/pulls/:number/reviews", prHandlers.ListReviews)
				repos.POST("/:owner/:repo/pulls/:number/comments", prHandlers.CreateReviewComment)
				repos.GET("/:owner/:repo/pulls/:number/comments", prHandlers.ListReviewComments)

				// Issue operations (require authentication)
				repos.POST("/:owner/:repo/issues", issueHandlers.CreateIssue)
				repos.PATCH("/:owner/:repo/issues/:number", issueHandlers.UpdateIssue)
				repos.POST("/:owner/:repo/issues/:number/close", issueHandlers.CloseIssue)
				repos.POST("/:owner/:repo/issues/:number/reopen", issueHandlers.ReopenIssue)
				repos.POST("/:owner/:repo/issues/:number/lock", issueHandlers.LockIssue)
				repos.POST("/:owner/:repo/issues/:number/unlock", issueHandlers.UnlockIssue)

				// Comment operations (require authentication)
				repos.POST("/:owner/:repo/issues/:number/comments", commentHandlers.CreateComment)
				repos.PATCH("/:owner/:repo/issues/:number/comments/:comment_id", commentHandlers.UpdateComment)
				repos.DELETE("/:owner/:repo/issues/:number/comments/:comment_id", commentHandlers.DeleteComment)

				// Label operations (require authentication)
				repos.POST("/:owner/:repo/labels", labelHandlers.CreateLabel)
				repos.PATCH("/:owner/:repo/labels/:name", labelHandlers.UpdateLabel)
				repos.DELETE("/:owner/:repo/labels/:name", labelHandlers.DeleteLabel)

				// Milestone operations (require authentication)
				repos.POST("/:owner/:repo/milestones", milestoneHandlers.CreateMilestone)
				repos.PATCH("/:owner/:repo/milestones/:number", milestoneHandlers.UpdateMilestone)
				repos.DELETE("/:owner/:repo/milestones/:number", milestoneHandlers.DeleteMilestone)
				repos.POST("/:owner/:repo/milestones/:number/close", milestoneHandlers.CloseMilestone)
				repos.POST("/:owner/:repo/milestones/:number/reopen", milestoneHandlers.ReopenMilestone)
			}

			// Admin-only operations
			adminRepos := protected.Group("/repositories")
			adminRepos.Use(middleware.AdminMiddleware())
			{
				adminRepos.DELETE("/:owner/:repo/issues/:number", issueHandlers.DeleteIssue)
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
			}
		}
	}
}
