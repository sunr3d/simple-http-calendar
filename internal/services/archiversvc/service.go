package archiversvc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/config"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/internal/interfaces/services"
)

var _ services.ArchiveService = (*archiveSvc)(nil)

type archiveSvc struct {
	repo     infra.Database
	logger   *zap.Logger
	interval time.Duration
}

// New - конструктор сервиса архивации.
func New(repo infra.Database, logger *zap.Logger, cfg config.ArchiverConfig) services.ArchiveService {
	return &archiveSvc{
		repo:     repo,
		logger:   logger,
		interval: cfg.Interval,
	}
}

// Start - запуск сервиса архивации.
// По таймеру проверяет и архивирует события, которые уже прошли.
func (s *archiveSvc) Start(ctx context.Context) error {
	logger := s.logger.With(
		zap.String("service", "archiver"),
		zap.String("op", "Start"),
	)

	logger.Info("запуск сервиса архивации...",
		zap.Duration("interval", s.interval),
	)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.archiveOldEvents(ctx); err != nil {
				logger.Warn("ошибка при архивации событий", zap.Error(err))
				continue
			}
		case <-ctx.Done():
			logger.Info("отмена контекста, сервис архивации остановлен")
			return ctx.Err()
		}
	}
}

// archiveOldEvents - архивирует события, которые уже прошли.
func (s *archiveSvc) archiveOldEvents(ctx context.Context) error {
	logger := s.logger.With(
		zap.String("service", "archiver"),
		zap.String("op", "archiveOldEvents"),
	)

	archived := false
	events, err := s.repo.List(ctx, &infra.ListOptions{
		Archived: &archived,
	})
	if err != nil {
		return fmt.Errorf("repo.List: %w", err)
	}

	now := time.Now()
	for _, event := range events {
		if event.Date.Before(now) {
			event.Archived = true
			if err := s.repo.Update(ctx, &event); err != nil {
				logger.Warn("ошибка при архивации события",
					zap.String("event_id", event.ID),
					zap.Error(err))
				continue
			}
			logger.Info("событие архивировано", zap.String("event_id", event.ID))
		}
	}

	return nil
}
