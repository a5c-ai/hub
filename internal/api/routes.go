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