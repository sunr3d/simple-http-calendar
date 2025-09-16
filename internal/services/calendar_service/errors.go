package calendar_service

import "errors"

var (
	errUserID     = errors.New("некорректный user_id")
	errEventID    = errors.New("некорректный event_id")
	errEmptyEvent = errors.New("описание события не может быть пустым")
)
