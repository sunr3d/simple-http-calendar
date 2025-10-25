package inmembroker

import (
	"context"
	"runtime/debug"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/models"
)

var _ infra.Broker = (*inmemBroker)(nil)

type inmemBroker struct {
	eventChan chan *models.Event
	logger    *zap.Logger
}

func New(chanSize int, logger *zap.Logger) infra.Broker {
	return &inmemBroker{
		eventChan: make(chan *models.Event, chanSize),
		logger:    logger,
	}
}

func (b *inmemBroker) Publish(ctx context.Context, event *models.Event) error {
	logger := b.logger.With(
		zap.String("service", "inmembroker"),
		zap.String("op", "Publish"),
	)

	select {
	case b.eventChan <- event:
		logger.Info("событие успешно отправлено в брокер",
			zap.String("event_id", event.ID),
			zap.Int64("user_id", event.UserID),
			zap.String("event", event.Text),
			zap.Time("date", event.Date),
		)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		logger.Warn("брокер переполнен, событие не может быть отправлено")
		return nil
	}
}

func (b *inmemBroker) Subscribe(ctx context.Context, handler func(ctx context.Context, event *models.Event) error) error {
	logger := b.logger.With(
		zap.String("service", "inmembroker"),
		zap.String("op", "Subscribe"),
	)
	logger.Info("запуск подписки на события брокера")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("паника в горутине подписки на события брокера",
					zap.Any("rec", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()

		for {
			select {
			case event := <-b.eventChan:
				logger.Info("получено событие из брокера",
					zap.String("event_id", event.ID),
				)
				if err := handler(ctx, event); err != nil {
					logger.Error("ошибка при обработке события",
						zap.Error(err),
						zap.String("event_id", event.ID),
					)
				}
			case <-ctx.Done():
				logger.Info("контекст завершен, завершение подписки на события брокера")
				return
			}
		}
	}()

	return nil
}
