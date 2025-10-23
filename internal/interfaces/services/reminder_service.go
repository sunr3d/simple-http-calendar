package services

import (
	"context"
	"time"
)

type ReminderService interface {
	Start(ctx context.Context, interval time.Duration) error
}
