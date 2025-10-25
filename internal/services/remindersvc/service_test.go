package remindersvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/infra/inmembroker"
	"github.com/sunr3d/simple-http-calendar/internal/infra/inmemdb"
	"github.com/sunr3d/simple-http-calendar/models"
)

func newReminderSvc(t *testing.T) *reminderSvc {
	t.Helper()

	logger := zap.NewNop()

	repo := inmemdb.New(logger)
	broker := inmembroker.New(100, logger)

	s := New(repo, broker, logger)
	rs, ok := s.(*reminderSvc)

	require.True(t, ok)
	return rs
}

func TestSendReminder(t *testing.T) {
	svc := newReminderSvc(t)

	event := &models.Event{
		ID:       "test-1",
		UserID:   1,
		Date:     time.Now().Add(-1 * time.Hour),
		Text:     "test event",
		Reminder: true,
	}

	svc.sendReminder(event)

	assert.True(t, true)
}

func TestCheckPendingReminders(t *testing.T) {
	svc := newReminderSvc(t)
	ctx := context.Background()

	pastEvent := models.Event{
		ID:           "past-1",
		UserID:       1,
		Date:         time.Now().Add(-1 * time.Hour),
		Text:         "past event",
		Reminder:     true,
		ReminderSent: false,
	}

	err := svc.repo.Create(ctx, &pastEvent)
	require.NoError(t, err)

	err = svc.checkPendingReminders(ctx)
	require.NoError(t, err)
}

func TestHandleReminder(t *testing.T) {
	svc := newReminderSvc(t)
	ctx := context.Background()

	pastEvent := &models.Event{
		ID:       "past-1",
		UserID:   1,
		Date:     time.Now().Add(-1 * time.Hour),
		Text:     "past event",
		Reminder: true,
	}

	err := svc.repo.Create(ctx, pastEvent)
	require.NoError(t, err)

	err = svc.handleReminder(ctx, pastEvent)
	require.NoError(t, err)

	updatedEvent, err := svc.repo.Read(ctx, pastEvent.ID)
	require.NoError(t, err)
	assert.True(t, updatedEvent.ReminderSent)
	assert.NotNil(t, updatedEvent.ReminderSentAt)
}

func TestHandleReminderFuture(t *testing.T) {
	svc := newReminderSvc(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	futureEvent := &models.Event{
		ID:       "future-1",
		UserID:   1,
		Date:     time.Now().Add(1 * time.Hour),
		Text:     "future event",
		Reminder: true,
	}

	err := svc.repo.Create(ctx, futureEvent)
	require.NoError(t, err)

	err = svc.handleReminder(ctx, futureEvent)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestHandleReminderUpdateError(t *testing.T) {
	svc := newReminderSvc(t)
	ctx := context.Background()

	pastEvent := &models.Event{
		ID:       "past-1",
		UserID:   1,
		Date:     time.Now().Add(-1 * time.Hour),
		Text:     "past event",
		Reminder: true,
	}

	err := svc.handleReminder(ctx, pastEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestStart(t *testing.T) {
	svc := newReminderSvc(t)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	pastEvent := models.Event{
		ID:           "past-1",
		UserID:       1,
		Date:         time.Now().Add(-1 * time.Hour),
		Text:         "old event",
		Reminder:     true,
		ReminderSent: false,
	}

	err := svc.repo.Create(ctx, &pastEvent)
	require.NoError(t, err)

	mockInterval := 50 * time.Millisecond
	err = svc.Start(ctx, mockInterval)
	require.Error(t, err)
	require.Equal(t, context.DeadlineExceeded, err)

	updatedEvent, err := svc.repo.Read(ctx, pastEvent.ID)
	require.NoError(t, err)
	assert.True(t, updatedEvent.ReminderSent)
}
