package inmemdb

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/sunr3d/simple-http-calendar/internal/interfaces/infra"
	"github.com/sunr3d/simple-http-calendar/models"
)

var _ infra.Database = (*inmemRepo)(nil)

type inmemRepo struct {
	data   map[string]models.Event
	logger *zap.Logger
	mu     sync.RWMutex
}

func New(log *zap.Logger) infra.Database {
	return &inmemRepo{
		data:   make(map[string]models.Event),
		logger: log,
	}
}

func (db *inmemRepo) Create(_ context.Context, event *models.Event) error {
	if event == nil {
		return errNilEvent
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[event.ID]; exists {
		return errDuplicate
	}

	db.data[event.ID] = *event
	return nil
}

func (db *inmemRepo) Read(_ context.Context, eventID string) (*models.Event, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	evnt, exists := db.data[eventID]
	if !exists {
		return nil, errNotFound
	}

	return &evnt, nil
}

func (db *inmemRepo) Update(_ context.Context, event *models.Event) error {
	if event == nil {
		return errNilEvent
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.data[event.ID]; !exists {
		return errNotFound
	}

	db.data[event.ID] = *event

	return nil
}

func (db *inmemRepo) Delete(_ context.Context, eventID string) (bool, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, exists := db.data[eventID]
	if !exists {
		return false, errNotFound
	}

	delete(db.data, eventID)

	return true, nil
}

func (db *inmemRepo) List(_ context.Context, opts *infra.ListOptions) ([]models.Event, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	res := make([]models.Event, 0, len(db.data))
	for _, evnt := range db.data {
		if opts != nil {
			if opts.UserID != nil && evnt.UserID != *opts.UserID {
				continue
			}
			if opts.Archived != nil && evnt.Archived != *opts.Archived {
				continue
			}

			if opts.From != nil {
				eventDay := time.Date(evnt.Date.Year(), evnt.Date.Month(), evnt.Date.Day(), 0, 0, 0, 0, time.Local)
				if eventDay.Before(*opts.From) {
					continue
				}
			}
			if opts.To != nil {
				eventDay := time.Date(evnt.Date.Year(), evnt.Date.Month(), evnt.Date.Day(), 0, 0, 0, 0, time.Local)
				if eventDay.After(*opts.To) {
					continue
				}
			}
		}

		res = append(res, evnt)
	}

	return res, nil
}
