package calendarsvc

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

func newSvc(t *testing.T) *calendarService {
	t.Helper()

	logger := zap.NewNop()

	repo := inmemdb.New(logger)
	broker := inmembroker.New(100, logger)

	s := New(repo, broker, logger)
	cs, ok := s.(*calendarService)

	require.True(t, ok)

	return cs
}

func TestCreate(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	day := time.Date(2025, 1, 2, 13, 14, 15, 0, time.UTC)

	id, err := svc.CreateEvent(ctx, models.Event{
		UserID: 1,
		Date:   day,
		Text:   "meeting",
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	events, err := svc.GetEventsForDay(ctx, 1, day)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "meeting", events[0].Text)
	assert.Equal(t, int64(1), events[0].UserID)
	assert.Equal(t, day, events[0].Date)
}

func TestCreateWithReminder(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	day := time.Date(2025, 1, 2, 13, 14, 15, 0, time.UTC)

	id, err := svc.CreateEvent(ctx, models.Event{
		UserID:   1,
		Date:     day,
		Text:     "meeting with reminder",
		Reminder: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)

	events, err := svc.GetEventsForDay(ctx, 1, day)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "meeting with reminder", events[0].Text)
	assert.Equal(t, int64(1), events[0].UserID)
	assert.Equal(t, day, events[0].Date)
	assert.True(t, events[0].Reminder)
	assert.False(t, events[0].ReminderSent)
}

func TestUpdate(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	day := time.Date(2025, 2, 10, 9, 0, 0, 0, time.UTC)

	id, err := svc.CreateEvent(ctx, models.Event{UserID: 42, Date: day, Text: "old"})
	require.NoError(t, err)

	err = svc.UpdateEvent(ctx, models.Event{
		ID:     id,
		UserID: 42,
		Date:   day,
		Text:   "new",
	})
	require.NoError(t, err)

	list, err := svc.GetEventsForDay(ctx, 42, day)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "new", list[0].Text)
}

func TestDelete(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	day := time.Date(2025, 3, 3, 0, 0, 0, 0, time.UTC)

	id, err := svc.CreateEvent(ctx, models.Event{UserID: 7, Date: day, Text: "to remove"})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteEvent(ctx, id))

	list, err := svc.GetEventsForDay(ctx, 7, day)
	require.NoError(t, err)
	assert.Len(t, list, 0)
}

func TestValidationErrors(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()
	day := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)

	_, err := svc.CreateEvent(ctx, models.Event{UserID: 0, Date: day, Text: "x"})
	require.Error(t, err)

	err = svc.UpdateEvent(ctx, models.Event{ID: "", UserID: 1, Date: day, Text: "x"})
	require.Error(t, err)

	err = svc.DeleteEvent(ctx, "")
	require.Error(t, err)
}

func TestGetEventsForWeek(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	monday := time.Date(2025, 1, 6, 10, 0, 0, 0, time.UTC)
	tuesday := time.Date(2025, 1, 7, 11, 0, 0, 0, time.UTC)
	nextMonday := time.Date(2025, 1, 13, 12, 0, 0, 0, time.UTC)

	id1, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: monday, Text: "monday event"})
	require.NoError(t, err)

	id2, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: tuesday, Text: "tuesday event"})
	require.NoError(t, err)

	id3, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: nextMonday, Text: "next week event"})
	require.NoError(t, err)

	events, err := svc.GetEventsForWeek(ctx, 1, monday)
	require.NoError(t, err)

	assert.Len(t, events, 2)

	eventIDs := []string{events[0].ID, events[1].ID}
	assert.ElementsMatch(t, eventIDs, []string{id1, id2})
	assert.NotContains(t, eventIDs, id3)
}

func TestGetEventsForMonth(t *testing.T) {
	svc := newSvc(t)
	ctx := context.Background()

	jan1 := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	jan31 := time.Date(2025, 1, 31, 11, 0, 0, 0, time.UTC)
	feb1 := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)

	id1, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: jan1, Text: "jan 15 event"})
	require.NoError(t, err)

	id2, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: jan31, Text: "jan 31 event"})
	require.NoError(t, err)

	id3, err := svc.CreateEvent(ctx, models.Event{UserID: 1, Date: feb1, Text: "feb event"})
	require.NoError(t, err)

	events, err := svc.GetEventsForMonth(ctx, 1, jan1)
	require.NoError(t, err)

	assert.Len(t, events, 2)

	eventIDs := []string{events[0].ID, events[1].ID}
	assert.ElementsMatch(t, eventIDs, []string{id1, id2})
	assert.NotContains(t, eventIDs, id3)
}
