package calendarsvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sunr3d/simple-http-calendar/internal/infra/inmem"
	"github.com/sunr3d/simple-http-calendar/models"
)

func newSvc(t *testing.T) *calendarService {
	t.Helper()

	repo := inmem.New(nil)

	s := New(repo, nil)
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
	assert.Equal(t, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), events[0].Date)
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
