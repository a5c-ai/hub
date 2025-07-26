package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/a5c-ai/hub/internal/models"
	"github.com/a5c-ai/hub/internal/services"
	"github.com/a5c-ai/hub/internal/testutil"
)

func TestAnalyticsService_RecordAndGetEvents(t *testing.T) {
	db := testutil.NewTestDB(t)
	// migrate AnalyticsEvent schema
	require.NoError(t, db.AutoMigrate(&models.AnalyticsEvent{}))

	logger := logrus.New()
	svc := services.NewAnalyticsService(db, logger)

	now := time.Now().UTC()
	actorID := uuid.New()
	repoID := uuid.New()
	event := &models.AnalyticsEvent{
		EventType:    models.EventType("test_event"),
		ActorID:      &actorID,
		ActorType:    "user",
		TargetType:   "repository",
		TargetID:     &repoID,
		RepositoryID: &repoID,
		UserAgent:    "test-agent",
		IPAddress:    "127.0.0.1",
		SessionID:    "session-1",
		RequestID:    "req-1",
		Metadata:     `{"key":"value"}`,
		CreatedAt:    now,
		Status:       "success",
	}

	// record the event
	require.NoError(t, svc.RecordEvent(context.Background(), event))

	// retrieve events without filters
	events, total, err := svc.GetEvents(context.Background(), services.EventFilters{})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, events, 1)

	fetched := events[0]
	require.Equal(t, event.EventType, fetched.EventType)
	require.Equal(t, *event.ActorID, *fetched.ActorID)
	require.Equal(t, event.TargetType, fetched.TargetType)
	require.Equal(t, *event.RepositoryID, *fetched.RepositoryID)
}
