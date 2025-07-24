package api

import (
	"net/http"

	"github.com/a5c-ai/hub/internal/auth"
	"github.com/a5c-ai/hub/internal/config"
	"github.com/a5c-ai/hub/internal/controllers"
	"github.com/a5c-ai/hub/internal/db"
	"github.com/a5c-ai/hub/internal/middleware"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRoutes(router *gin.Engine, database *db.Database, logger *logrus.Logger) {
	cfg, _ := config.Load()
	jwtManager := auth.NewJWTManager(cfg.JWT)

	// Initialize services
	activityService := services.NewActivityService(database.DB)
	orgService := services.NewOrganizationService(database.DB, activityService)
	memberService := services.NewMembershipService(database.DB, activityService)
	invitationService := services.NewInvitationService(database.DB, activityService)
	teamService := services.NewTeamService(database.DB, activityService)
	teamMembershipService := services.NewTeamMembershipService(database.DB, activityService)
	permissionService := services.NewPermissionService(database.DB, activityService)

	// Initialize controllers
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

	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		auth := v1.Group("/auth")
		{
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{"message": "Login endpoint - to be implemented"})
			})
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{"message": "Register endpoint - to be implemented"})
			})
			auth.POST("/logout", func(c *gin.Context) {
				c.JSON(http.StatusNotImplemented, gin.H{"message": "Logout endpoint - to be implemented"})
			})
		}

		// Public invitation acceptance endpoint
		v1.POST("/invitations/accept", orgController.AcceptInvitation)

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

			// User's organizations
			protected.GET("/user/organizations", orgController.GetUserOrganizations)

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