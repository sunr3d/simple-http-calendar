package calendarsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/services"
	"github.com/sunr3d/simple-http-calendar/models"
)

var _ services.CalendarService = (*calendarService)(nil)

type calendarService struct {
	repo   infra.Database
	broker infra.Broker
	logger *zap.Logger
}

// New - конструктор сервиса календаря.
func New(repo infra.Database, broker infra.Broker, logger *zap.Logger) services.CalendarService {
	return &calendarService{
		repo:   repo,
		broker: broker,
		logger: logger,
	}
}

// CreateEvent - создает новое событие в календаре.
// Если событие имеет флаг напоминания, то оно отправляется в брокер для дальнейшей обработки.
func (s *calendarService) CreateEvent(ctx context.Context, event models.Event) (string, error) {
	if event.UserID <= 0 {
		return "", errUserID
	}
	if event.Text == "" {
		return "", errEmptyEvent
	}

	id := uuid.NewString()
	newEvent := &models.Event{
		ID:       id,
		UserID:   event.UserID,
		Date:     event.Date,
		Text:     event.Text,
		Reminder: event.Reminder,
	}

	if err := s.repo.Create(ctx, newEvent); err != nil {
		return "", fmt.Errorf("repo.Create: %w", err)
	}

	if newEvent.Reminder {
		if err := s.broker.Publish(ctx, newEvent); err != nil {
			return id, fmt.Errorf("broker.Publish: %w", err)
		}
	}

	return id, nil
}

// UpdateEvent - обновляет событие в календаре.
// Если событие имеет флаг напоминания, то оно отправляется в брокер для дальнейшей обработки.
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

	data, err := s.repo.Read(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("repo.Read: %w", err)
	}

	data.Date = event.Date
	data.Text = event.Text
	if event.Reminder {
		data.Reminder = true
		if err := s.broker.Publish(ctx, data); err != nil {
			return fmt.Errorf("broker.Publish: %w", err)
		}
	}

	return s.repo.Update(ctx, data)
}

// DeleteEvent - удаляет событие из календара.
func (s *calendarService) DeleteEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		return errEventID
	}

	_, err := s.repo.Delete(ctx, eventID)

	return err
}

// GetEventsForDay - получает все события для указанного дня.
func (s *calendarService) GetEventsForDay(
	ctx context.Context,
	userID int64,
	dateRange time.Time,
) ([]models.Event, error) {
	if userID <= 0 {
		return nil, errUserID
	}

	day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.Local)
	archived := false

	return s.repo.List(
		ctx,
		&infra.ListOptions{
			UserID:   &userID,
			Archived: &archived,
			From:     &day,
			To:       &day,
		},
	)
}

// GetEventsForWeek - получает все события для указанной недели.
func (s *calendarService) GetEventsForWeek(
	ctx context.Context,
	userID int64,
	dateRange time.Time,
) ([]models.Event, error) {
	if userID <= 0 {
		return nil, errUserID
	}

	day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.Local)
	weekday := int(day.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := day.AddDate(0, 0, -(weekday - 1))
	weekEnd := weekStart.AddDate(0, 0, 6)
	archived := false

	return s.repo.List(
		ctx,
		&infra.ListOptions{
			UserID:   &userID,
			Archived: &archived,
			From:     &weekStart,
			To:       &weekEnd,
		},
	)
}

// GetEventsForMonth - получает все события для указанного месяца.
func (s *calendarService) GetEventsForMonth(
	ctx context.Context,
	userID int64,
	dateRange time.Time,
) ([]models.Event, error) {
	if userID <= 0 {
		return nil, errUserID
	}

	day := time.Date(dateRange.Year(), dateRange.Month(), dateRange.Day(), 0, 0, 0, 0, time.Local)
	monthStart := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, -1)
	archived := false

	return s.repo.List(
		ctx,
		&infra.ListOptions{
			UserID:   &userID,
			Archived: &archived,
			From:     &monthStart,
			To:       &monthEnd,
		},
	)
}
