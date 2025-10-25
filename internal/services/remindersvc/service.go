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
	s.logger.Info("запуск сервиса напоминаний...",
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
				s.logger.Warn("ошибка при проверке ожидающих напоминаний", zap.Error(err))
				continue
			}
		case <-ctx.Done():
			s.logger.Info("отмена контекста, сервис напоминаний остановлен")
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

	s.sendReminder(event)
	event.ReminderSent = true
	sentAt := time.Now()
	event.ReminderSentAt = &sentAt

	if err := s.repo.Update(ctx, event); err != nil {
		return fmt.Errorf("repo.Update: %w", err)
	}

	return nil
}

// checkPendingReminders - проверяет ожидающие напоминания из БД (фоллбэк хелпер).
// Получает все события из БД и проверяет, если событие уже в прошлом, то оно отправляется сразу.
// Если событие еще не наступило, то ждет и отправляет позже.
func (s *reminderSvc) checkPendingReminders(ctx context.Context) error {
	events, err := s.repo.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("repo.List: %w", err)
	}

	now := time.Now()
	for _, event := range events {
		if event.Reminder &&
			!event.ReminderSent &&
			(now.After(event.Date) || now.Equal(event.Date)) {
			go func(e models.Event) {
				s.sendReminder(&e)

				e.ReminderSent = true
				sentAt := time.Now()
				e.ReminderSentAt = &sentAt

				if err := s.repo.Update(ctx, &e); err != nil {
					s.logger.Warn("ошибка при обновлении статуса напоминания в БД", zap.Error(err))
				}
			}(event)
		}
	}

	return nil
}

// sendReminder - отправляет напоминание.
func (s *reminderSvc) sendReminder(event *models.Event) {
	s.logger.Info("отправлено напоминание",
		zap.Int64("user_id", event.UserID),
		zap.String("event_id", event.ID),
		zap.String("event", event.Text),
		zap.Time("date", event.Date),
	)
	fmt.Printf("НАПОМИНАНИЕ: событие '%s' начинается сейчас!\n", event.Text)
}
