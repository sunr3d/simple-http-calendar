package infra

import (
	"context"
	"time"

	"github.com/sunr3d/simple-http-calendar/models"
)

type ListOptions struct {
	UserID       *int64
	Archived     *bool
	ReminderSent *bool
	From         *time.Time
	To           *time.Time
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.2 --name=Database --output=../../../mocks --filename=mock_database.go --with-expecter
type Database interface {
	Create(ctx context.Context, event *models.Event) error
	Read(ctx context.Context, eventID string) (*models.Event, error)
	Update(ctx context.Context, event *models.Event) error
	Delete(ctx context.Context, eventID string) (bool, error)
	List(ctx context.Context, opts *ListOptions) ([]models.Event, error)
}
