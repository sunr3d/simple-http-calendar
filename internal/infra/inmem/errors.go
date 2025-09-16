package inmem

import "errors"

var (
	errDuplicate = errors.New("запись с таким ID уже существует")
	errNotFound  = errors.New("запись не найдена")
	errNilEvent  = errors.New("event не может быть nil")
)
