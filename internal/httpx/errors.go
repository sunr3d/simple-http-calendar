package httpx

import "errors"

var (
	ErrJSONMarshal = errors.New("не удалось сериализовать JSON")
	ErrWriteBody   = errors.New("не удалось записать тело ответа")
)