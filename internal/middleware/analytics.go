package middleware

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AnalyticsMiddleware creates middleware for automatic analytics data collection
func AnalyticsMiddleware(analyticsService services.AnalyticsService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timing the request
		startTime := time.Now()
		
		// Generate request ID if not present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Header("X-Request-ID", requestID)
		}

		// Process the request
		c.Next()

		// Collect analytics data after request completion
		go func() {
			duration := time.Since(startTime)
			collectAnalyticsData(c, analyticsService, logger, requestID, duration)
		}()
	}
}

// collectAnalyticsData collects and records analytics data for the request
func collectAnalyticsData(c *gin.Context, analyticsService services.AnalyticsService, logger *logrus.Logger, requestID string, duration time.Duration) {
	ctx := context.Background()
	
	// Extract request information
	method := c.Request.Method
	path := c.Request.URL.Path
	statusCode := c.Writer.Status()
	responseSize := int64(c.Writer.Size())
	userAgent := c.GetHeader("User-Agent")
	ipAddress := getClientIP(c)
	
	// Get user information if authenticated
	var actorID *uuid.UUID
	var sessionID string
	if userIDInterface, exists := c.Get("user_id"); exists {
		if uid, ok := parseUserID(userIDInterface); ok {
			actorID = &uid
		}
	}
	if sessionIDInterface, exists := c.Get("session_id"); exists {
		if sid, ok := sessionIDInterface.(string); ok {
			sessionID = sid
		}
	}

	// Extract repository and organization context from path
	var repositoryID, organizationID *uuid.UUID
	if repoID, orgID := extractRepositoryContext(path); repoID != nil || orgID != nil {
		repositoryID = repoID
		organizationID = orgID
	}

	// Record performance log
	performanceLog := &models.PerformanceLog{
		RequestID:      requestID,
		Method:         method,
		Path:           path,
		StatusCode:     statusCode,
		Duration:       duration.Milliseconds(),
		ResponseSize:   responseSize,
		UserID:         actorID,
		IPAddress:      ipAddress,
		UserAgent:      userAgent,
		RepositoryID:   repositoryID,
		OrganizationID: organizationID,
	}

	// Add error information for failed requests
	if statusCode >= 400 {
		if errorMessage, exists := c.Get("error_message"); exists {
			if msg, ok := errorMessage.(string); ok {
				performanceLog.ErrorMessage = msg
			}
		}
		if stackTrace, exists := c.Get("stack_trace"); exists {
			if trace, ok := stackTrace.(string); ok {
				performanceLog.StackTrace = trace
			}
		}
	}

	if err := analyticsService.RecordPerformanceLog(ctx, performanceLog); err != nil {
		logger.WithError(err).Warn("Failed to record performance log")
	}

	// Record analytics event based on the request
	if event := createAnalyticsEvent(c, requestID, actorID, repositoryID, organizationID, ipAddress, userAgent, sessionID, statusCode); event != nil {
		if err := analyticsService.RecordEvent(ctx, event); err != nil {
			logger.WithError(err).Warn("Failed to record analytics event")
		}
	}

	// Record metrics for API usage
	recordAPIMetrics(ctx, analyticsService, method, path, statusCode, duration, repositoryID, organizationID, actorID)
}

// createAnalyticsEvent creates an analytics event based on the request
func createAnalyticsEvent(c *gin.Context, requestID string, actorID, repositoryID, organizationID *uuid.UUID, ipAddress, userAgent, sessionID string, statusCode int) *models.AnalyticsEvent {
	method := c.Request.Method
	path := c.Request.URL.Path
	
	var eventType models.EventType
	var targetType string
	var targetID *uuid.UUID
	var metadata map[string]interface{}

	// Determine event type based on path and method
	switch {
	// User authentication events
	case strings.HasPrefix(path, "/api/v1/auth/login") && method == "POST":
		eventType = models.EventUserLogin
		targetType = "user"
		targetID = actorID
		
	case strings.HasPrefix(path, "/api/v1/auth/logout") && method == "POST":
		eventType = models.EventUserLogout
		targetType = "user"
		targetID = actorID
		
	case strings.HasPrefix(path, "/api/v1/auth/register") && method == "POST":
		eventType = models.EventUserRegistration
		targetType = "user"
		targetID = actorID
		
	// Repository events
	case strings.Contains(path, "/repositories/") && strings.HasSuffix(path, ".git/git-receive-pack") && method == "POST":
		eventType = models.EventRepositoryPush
		targetType = "repository"
		targetID = repositoryID
		
	case strings.Contains(path, "/repositories/") && strings.HasSuffix(path, ".git/git-upload-pack") && method == "POST":
		eventType = models.EventRepositoryClone
		targetType = "repository"
		targetID = repositoryID
		
	case strings.HasPrefix(path, "/api/v1/repositories") && method == "POST":
		eventType = models.EventRepositoryCreated
		targetType = "repository"
		targetID = repositoryID
		
	case strings.Contains(path, "/repositories/") && method == "DELETE":
		eventType = models.EventRepositoryDeleted
		targetType = "repository"
		targetID = repositoryID
		
	case strings.Contains(path, "/pulls") && method == "POST":
		eventType = models.EventRepositoryPullRequest
		targetType = "repository"
		targetID = repositoryID
		metadata = map[string]interface{}{"action": "created"}
		
	case strings.Contains(path, "/issues") && method == "POST":
		eventType = models.EventRepositoryIssue
		targetType = "repository"
		targetID = repositoryID
		metadata = map[string]interface{}{"action": "created"}
		
	// Organization events
	case strings.HasPrefix(path, "/api/v1/organizations") && method == "POST":
		eventType = models.EventOrgCreated
		targetType = "organization"
		targetID = organizationID
		
	case strings.Contains(path, "/organizations/") && strings.Contains(path, "/members/") && method == "PUT":
		eventType = models.EventOrgMemberAdded
		targetType = "organization"
		targetID = organizationID
		
	case strings.Contains(path, "/organizations/") && strings.Contains(path, "/members/") && method == "DELETE":
		eventType = models.EventOrgMemberRemoved
		targetType = "organization"
		targetID = organizationID
		
	// API calls for analytics
	case strings.HasPrefix(path, "/api/v1/") && method == "GET":
		eventType = models.EventAPICall
		targetType = "api"
		metadata = map[string]interface{}{"endpoint": path, "method": method}
		
	// Page views (for non-API requests)
	case !strings.HasPrefix(path, "/api/") && method == "GET":
		eventType = models.EventPageView
		targetType = "page"
		metadata = map[string]interface{}{"page": path}
		
	default:
		// Don't create event for untracked requests
		return nil
	}

	// Determine status
	status := "success"
	if statusCode >= 400 && statusCode < 500 {
		status = "error"
	} else if statusCode >= 500 {
		status = "error"
	}

	// Convert metadata to JSON string
	var metadataJSON string
	if metadata != nil {
		if jsonBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(jsonBytes)
		}
	}

	actorType := "anonymous"
	if actorID != nil {
		actorType = "user"
	}

	return &models.AnalyticsEvent{
		EventType:      eventType,
		ActorID:        actorID,
		ActorType:      actorType,
		TargetType:     targetType,
		TargetID:       targetID,
		RepositoryID:   repositoryID,
		OrganizationID: organizationID,
		UserAgent:      userAgent,
		IPAddress:      ipAddress,
		SessionID:      sessionID,
		RequestID:      requestID,
		Metadata:       metadataJSON,
		Status:         status,
	}
}

// recordAPIMetrics records metrics for API usage
func recordAPIMetrics(ctx context.Context, analyticsService services.AnalyticsService, method, path string, statusCode int, duration time.Duration, repositoryID, organizationID, userID *uuid.UUID) {
	timestamp := time.Now()
	
	// Record response time metric
	responseTimeMetric := &models.AnalyticsMetric{
		Name:           "api_response_time",
		MetricType:     models.MetricTypeHistogram,
		Value:          float64(duration.Milliseconds()),
		Timestamp:      timestamp,
		RepositoryID:   repositoryID,
		OrganizationID: organizationID,
		UserID:         userID,
		Period:         "hourly",
		Tags:           `{"method":"` + method + `","path":"` + path + `","status":"` + string(rune(statusCode)) + `"}`,
	}
	
	// Record request count metric
	requestCountMetric := &models.AnalyticsMetric{
		Name:           "api_request_count",
		MetricType:     models.MetricTypeCounter,
		Value:          1,
		Timestamp:      timestamp,
		RepositoryID:   repositoryID,
		OrganizationID: organizationID,
		UserID:         userID,
		Period:         "hourly",
		Tags:           `{"method":"` + method + `","path":"` + path + `","status":"` + string(rune(statusCode)) + `"}`,
	}

	// Record error rate if applicable
	if statusCode >= 400 {
		errorRateMetric := &models.AnalyticsMetric{
			Name:           "api_error_count",
			MetricType:     models.MetricTypeCounter,
			Value:          1,
			Timestamp:      timestamp,
			RepositoryID:   repositoryID,
			OrganizationID: organizationID,
			UserID:         userID,
			Period:         "hourly",
			Tags:           `{"method":"` + method + `","path":"` + path + `","status":"` + string(rune(statusCode)) + `"}`,
		}
		
		go analyticsService.RecordMetric(ctx, errorRateMetric)
	}

	// Record metrics asynchronously
	go analyticsService.RecordMetric(ctx, responseTimeMetric)
	go analyticsService.RecordMetric(ctx, requestCountMetric)
}

// Helper functions

func getClientIP(c *gin.Context) string {
	// Check various headers for the real client IP
	clientIP := c.GetHeader("X-Forwarded-For")
	if clientIP != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if strings.Contains(clientIP, ",") {
			clientIP = strings.Split(clientIP, ",")[0]
		}
		return strings.TrimSpace(clientIP)
	}
	
	clientIP = c.GetHeader("X-Real-IP")
	if clientIP != "" {
		return clientIP
	}
	
	clientIP = c.GetHeader("X-Client-IP")
	if clientIP != "" {
		return clientIP
	}
	
	return c.ClientIP()
}

func parseUserID(userID interface{}) (uuid.UUID, bool) {
	switch v := userID.(type) {
	case uuid.UUID:
		return v, true
	case string:
		if uid, err := uuid.Parse(v); err == nil {
			return uid, true
		}
	}
	return uuid.Nil, false
}

func extractRepositoryContext(path string) (*uuid.UUID, *uuid.UUID) {
	// This is a simplified version - in a real implementation,
	// you would need to parse the owner/repo from the path and 
	// look up the actual repository and organization IDs from the database
	
	// For now, return nil - this would need to be implemented 
	// with proper path parsing and database lookups
	return nil, nil
}