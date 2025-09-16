package validators

import (
	"strings"

	"github.com/sunr3d/simple-http-calendar/models"
)

func ValidateCreatePayload(payload models.Event) error {
	if payload.UserID <= 0 {
		return ErrBadUserID
	}
	if strings.TrimSpace(payload.Text) == "" {
		return ErrBadEventText
	}
	if payload.Date.IsZero() {
		return ErrBadDate
	}

	return nil
}

func ValidateUpdate(payload models.Event) error {
	if strings.TrimSpace(payload.ID) == "" {
		return ErrBadEventID
	}

	return ValidateCreatePayload(payload)
}

func ValidateFilter(filter models.EventsByDay) error {
	if filter.UserID <= 0 {
		return ErrBadUserID
	}
	if filter.Day.IsZero() {
		return ErrBadDate
	}

	return nil
}

func ValidateDelete(id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrBadEventID
	}

	return nil
}
