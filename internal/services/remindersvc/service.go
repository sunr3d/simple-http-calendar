package remindersvc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/services"
	"github.com/sunr3d/simple-http-calendar/models"
)

var _ services.ReminderService = (*reminderSvc)(nil)

type reminderSvc struct {
	repo   infra.Database
	broker infra.Broker
	logger *zap.Logger
}

// New - конструктор сервиса напоминаний.
func New(repo infra.Database, broker infra.Broker, logger *zap.Logger) services.ReminderService {
	return &reminderSvc{
		repo:   repo,
		broker: broker,
		logger: logger,
	}
}

// Start - запуск сервиса напоминаний.
// Подписывается на канал событий брокера и проверяет ожидающие напоминания.
// Есть фоллбэк на случай, если событие уже в прошлом, то оно отправляется сразу.
func (s *reminderSvc) Start(ctx context.Context, interval time.Duration) error {
	logger := s.logger.With(
		zap.String("service", "reminder"),
		zap.String("op", "Start"),
	)
	logger.Info("запуск сервиса напоминаний...",
		zap.Duration("interval", interval),
	)

	if err := s.broker.Subscribe(ctx, s.handleReminder); err != nil {
		return fmt.Errorf("broker.Subscribe: %w", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.checkPendingReminders(ctx); err != nil {
				logger.Warn("ошибка при проверке ожидающих напоминаний", zap.Error(err))
				continue
			}
		case <-ctx.Done():
			logger.Info("отмена контекста, сервис напоминаний остановлен")
			return ctx.Err()
		}
	}
}

// handleReminder - обработчик событий брокера.
// Проверяет, если событие уже в прошлом, то оно отправляется сразу.
// Если событие еще не наступило, то ждет и отправляет позже.
func (s *reminderSvc) handleReminder(ctx context.Context, event *models.Event) error {
	waitDur := time.Until(event.Date)
	if waitDur > 0 {
		select {
		case <-time.After(waitDur):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.sendReminder(ctx, event)

	event.ReminderSent = true
	sentAt := time.Now()
	event.ReminderSentAt = &sentAt

	if err := s.repo.Update(ctx, event); err != nil {
		return fmt.Errorf("repo.Update: %w", err)
	}

	return nil
}

// checkPendingReminders - проверяет ожидающие напоминания из БД (фоллбэк хелпер).
// Получает из БД все события, у которых не был отправлен статус напоминания.
// Если событие уже в прошлом или сейчас время совпадает с временем события,
// то напоминание отправляется сразу, а статус отправки напоминания устанавливается в true.
func (s *reminderSvc) checkPendingReminders(ctx context.Context) error {
	logger := s.logger.With(
		zap.String("service", "reminder"),
		zap.String("op", "checkPendingReminders"),
	)

	reminderSent := false
	events, err := s.repo.List(ctx, &infra.ListOptions{
		ReminderSent: &reminderSent,
	})
	if err != nil {
		return fmt.Errorf("repo.List: %w", err)
	}

	now := time.Now()
	for _, event := range events {
		if event.Reminder &&
			!event.ReminderSent &&
			(now.After(event.Date) || now.Equal(event.Date)) {
			event.ReminderSent = true
			sentAt := time.Now()
			event.ReminderSentAt = &sentAt

			if err := s.repo.Update(ctx, &event); err != nil {
				logger.Warn("ошибка при обновлении статуса напоминания в БД", zap.Error(err))
				continue
			}

			go s.sendReminder(ctx, &event)
		}
	}

	return nil
}

// sendReminder - отправляет напоминание.
func (s *reminderSvc) sendReminder(ctx context.Context, event *models.Event) {
	logger := s.logger.With(
		zap.String("service", "reminder"),
		zap.String("op", "sendReminder"),
	)

	select {
	case <-ctx.Done():
		return
	default:
	}

	logger.Info("отправлено напоминание",
		zap.Int64("user_id", event.UserID),
		zap.String("event_id", event.ID),
		zap.String("event", event.Text),
		zap.Time("date", event.Date),
	)
	fmt.Printf("НАПОМИНАНИЕ: событие '%s' начинается сейчас!\n", event.Text)
}
