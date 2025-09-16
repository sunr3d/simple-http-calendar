package infra

import (
	"context"
	"time"

	"github.com/sunr3d/simple-http-calendar/models"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}

type Database interface {
	Create(ctx context.Context, event *models.Event) error
	Read(ctx context.Context, eventID string) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, eventID string) (bool, error)

	ListByTimeRange(ctx context.Context, userID int64, tr TimeRange) ([]models.Event, error)
}
