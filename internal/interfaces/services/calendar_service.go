package services

import (
	"context"
	"time"

	"github.com/sunr3d/simple-http-calendar/models"
)

type CalendarService interface {
	CreateEvent(ctx context.Context, event models.Event) (string, error)
	UpdateEvent(ctx context.Context, event models.Event) error
	DeleteEvent(ctx context.Context, eventID string) error

	GetEventsForDay(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error)
	GetEventsForWeek(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error)
	GetEventsForMonth(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error)
}
