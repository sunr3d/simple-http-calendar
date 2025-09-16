package calendar_service

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/services"
	"github.com/sunr3d/simple-http-calendar/models"
)

var _ services.CalendarService = (*calendarService)(nil)

type calendarService struct {
	repo   infra.Database
	logger *zap.Logger
}

func New(repo infra.Database, logger *zap.Logger) services.CalendarService {
	return &calendarService{
		repo:   repo,
		logger: logger,
	}
}

func (s *calendarService) CreateEvent(ctx context.Context, event models.Event) (string, error) {
	if event.UserID <= 0 {
		return "", errUserID
	}
	if event.Text == "" {
		return "", errEventID
	}

	d := time.Date(event.Date.Year(), event.Date.Month(), event.Date.Day(), 0, 0, 0, 0, time.UTC)

	id := uuid.NewString()
	newEvent := &models.Event{
		ID:     id,
		UserID: event.UserID,
		Date:   d,
		Text:   event.Text,
	}

	if err := s.repo.Create(ctx, newEvent); err != nil {
		return "", fmt.Errorf("repo.Create: %w", err)
	}

    return id, nil
}

func (s *calendarService) UpdateEvent(ctx context.Context, event models.Event) error {
    if event.ID == "" {
        return errEventID
    }
    if event.UserID <= 0 {
        return errUserID
    }
    if event.Text == "" {
        return errEmptyEvent
    }

    d := time.Date(event.Date.Year(), event.Date.Month(), event.Date.Day(), 0, 0, 0, 0, time.UTC)

    data, err := s.repo.Read(ctx, event.ID)
    if err != nil {
        return fmt.Errorf("repo.Read: %w", err)
    }

    data.UserID = event.UserID
    data.Date = d
    data.Text = event.Text
    
    return s.repo.Update(ctx, data)
}

func (s *calendarService) DeleteEvent(ctx context.Context, eventID string) error {
    if eventID == "" {
        return errEventID
    }

    _, err := s.repo.Delete(ctx, eventID)

    return err
}

func (s *calendarService) GetEventsForDay(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error) {
    if userID <= 0 {
        return nil, errUserID
    }

    day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.UTC)
    tr := infra.TimeRange{From: day, To: day}

    return s.repo.ListByTimeRange(ctx, userID, tr)
}

func (s *calendarService) GetEventsForWeek(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error) {
    if userID <= 0 {
        return nil, errUserID
    }

    day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.UTC)
    
    weekday := int(day.Weekday())
    if weekday == 0 {
        weekday = 7
    }
    weekStart := day.AddDate(0, 0, -(weekday-1))
    weekEnd := weekStart.AddDate(0, 0, 6)

    tr := infra.TimeRange{From: weekStart, To: weekEnd}

    return s.repo.ListByTimeRange(ctx, userID, tr)
}

func (s *calendarService) GetEventsForMonth(ctx context.Context, userID int64, dateRange time.Time) ([]models.Event, error) {
    if userID <= 0 {
        return nil, errUserID
    }

    day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.UTC)

    monthStart := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
    monthEnd := monthStart.AddDate(0, 1, -1)

    tr := infra.TimeRange{From: monthStart, To: monthEnd}

    return s.repo.ListByTimeRange(ctx, userID, tr)
}
