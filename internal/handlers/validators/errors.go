package validators

import "errors"

var (
	ErrBadUserID    = errors.New("некорректный user_id")
	ErrBadEventID   = errors.New("некорректный event_id")
	ErrBadDate      = errors.New("некорректная дата, ожидается YYYY-MM-DD")
	ErrBadEventText = errors.New("текст события не может быть пустым")
)
