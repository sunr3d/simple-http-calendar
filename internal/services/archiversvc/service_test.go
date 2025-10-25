package archiversvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmemdb"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/models"
)

func newArchiveSvc(t *testing.T) *archiveSvc {
	t.Helper()

	logger := zap.NewNop()
	repo := inmemdb.New(logger)
	cfg := config.ArchiverConfig{
		Interval: 1 * time.Minute,
	}

	s := New(repo, logger, cfg)
	as, ok := s.(*archiveSvc)

	require.True(t, ok)
	return as
}

func TestArchiveOldEvents(t *testing.T) {
	svc := newArchiveSvc(t)
	ctx := context.Background()

	pastEvent := models.Event{
		ID:       "past-1",
		UserID:   1,
		Date:     time.Now().Add(-1 * time.Hour),
		Text:     "old event",
		Archived: false,
	}

	futureEvent := models.Event{
		ID:       "future-1",
		UserID:   1,
		Date:     time.Now().Add(1 * time.Hour),
		Text:     "future event",
		Archived: false,
	}

	err := svc.repo.Create(ctx, &pastEvent)
	require.NoError(t, err)
	err = svc.repo.Create(ctx, &futureEvent)
	require.NoError(t, err)

	err = svc.archiveOldEvents(ctx)
	require.NoError(t, err)

	archived := false
	events, err := svc.repo.List(ctx, &infra.ListOptions{Archived: &archived})
	require.NoError(t, err)

	assert.Len(t, events, 1)
	assert.Equal(t, "future-1", events[0].ID)
}

func TestStart(t *testing.T) {
	svc := newArchiveSvc(t)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pastEvent := models.Event{
		ID:       "past-1",
		UserID:   1,
		Date:     time.Now().Add(-1 * time.Hour),
		Text:     "old event",
		Archived: false,
	}

	err := svc.repo.Create(ctx, &pastEvent)
	require.NoError(t, err)

	svc.interval = 50 * time.Millisecond

	err = svc.Start(ctx)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)

	archived := false
	events, err := svc.repo.List(ctx, &infra.ListOptions{Archived: &archived})
	require.NoError(t, err)
	assert.Len(t, events, 0)
}
